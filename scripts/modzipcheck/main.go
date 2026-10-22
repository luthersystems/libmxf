// Copyright © 2026 Luther Systems, Ltd. All right reserved.

// Command modzipcheck builds the libmxf module zip from git, the way go get
// and the Go module proxy do, and fails if a Go source file embeds a Git LFS
// pointer.
//
// It exists because of libmxf#12: decode-go embedded an LFS-tracked wasm,
// CI checked out with LFS and passed, and every consumer got a 130-byte
// pointer. A test inside the module sees the checkout, not the zip; this
// command sees the zip. Pointers no Go code embeds (the JavaScript app.wasm,
// which npm publishes from a full checkout) are reported but do not fail.
//
// It is its own module, so the libmxf module does not depend on
// golang.org/x/mod, and the nested go.mod keeps this directory out of the
// libmxf zip.
//
// Usage: go run . <repo root> <git revision>
package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"maps"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"
)

const modulePath = "github.com/luthersystems/libmxf"

// lfsPointerPrefix is the first line of a Git LFS pointer file. decode-go's
// guestWasm checks the same prefix; keep the two in step. They are separate
// modules, so they cannot share it.
var lfsPointerPrefix = []byte("version https://git-lfs.github.com/spec/")

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: modzipcheck <repo root> <git revision>")
		os.Exit(2)
	}
	if err := check(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "modzipcheck:", err)
		os.Exit(1)
	}
}

func check(repoRoot, revision string) error {
	// CreateFromVCS runs git archive, which applies the LFS smudge filter
	// where git-lfs is installed without --skip, and so would put the real
	// content back in the zip. The proxy never smudges; neither may we.
	if err := os.Setenv("GIT_LFS_SKIP_SMUDGE", "1"); err != nil {
		return err
	}
	// CreateFromVCS hides git's error behind "could not find a recognized
	// version control system". Ask git first, so a failure (in CI, most
	// often "dubious ownership") says what git said.
	if out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--git-dir").CombinedOutput(); err != nil { //nolint:gosec // G204: fixed git command; repoRoot is the operator's own checkout
		return fmt.Errorf("git rev-parse in %s: %w: %s", repoRoot, err, bytes.TrimSpace(out))
	}
	// The version only names the zip's top directory; any valid one works.
	m := module.Version{Path: modulePath, Version: "v0.0.0-modzipcheck"}
	var buf bytes.Buffer
	if err := modzip.CreateFromVCS(&buf, m, repoRoot, revision, ""); err != nil {
		return fmt.Errorf("build module zip: %w", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}
	prefix := m.Path + "@" + m.Version + "/"
	files := map[string][]byte{}
	for _, f := range zr.File {
		b, err := readZipFile(f)
		if err != nil {
			return err
		}
		files[strings.TrimPrefix(f.Name, prefix)] = b
	}
	fmt.Printf("module zip of %s: %d bytes, %d files\n", revision, buf.Len(), len(files))

	names := slices.Sorted(maps.Keys(files))
	var bad []string
	embedded := map[string]bool{}
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		patterns, err := embedPatterns(name, files[name])
		if err != nil {
			return err
		}
		for _, pattern := range patterns {
			target := path.Join(path.Dir(name), pattern)
			embedded[target] = true
			content, ok := files[target]
			switch {
			case !ok:
				bad = append(bad, fmt.Sprintf("%s embeds %s, which is not a file in the zip", name, target))
			case bytes.HasPrefix(content, lfsPointerPrefix):
				bad = append(bad, fmt.Sprintf("%s embeds %s, which is a Git LFS pointer in the zip", name, target))
			default:
				if problem := checkContent(target, content); problem != "" {
					bad = append(bad, fmt.Sprintf("%s embeds %s, %s", name, target, problem))
					continue
				}
				fmt.Printf("ok: %s embeds %s (%d bytes)\n", name, target, len(content))
			}
		}
	}
	for _, name := range names {
		if !embedded[name] && bytes.HasPrefix(files[name], lfsPointerPrefix) {
			fmt.Printf("note: %s is a Git LFS pointer in the zip; no Go code embeds it\n", name)
		}
	}
	if len(bad) > 0 {
		return fmt.Errorf("%d bad embed(s):\n  %s", len(bad), strings.Join(bad, "\n  "))
	}
	return nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// embedPatterns returns the patterns of the //go:embed directives in a Go
// file. It reads comments with go/parser, so a //go:embed line inside a
// string literal or a block comment does not count. Patterns are split on
// spaces; libmxf uses no quoted patterns, globs or directories, and a pattern
// that names no file fails the check, so a new form cannot pass unchecked.
func embedPatterns(name string, src []byte) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	var patterns []string
	for _, group := range f.Comments {
		for _, c := range group.List {
			if rest, ok := strings.CutPrefix(c.Text, "//go:embed "); ok {
				patterns = append(patterns, strings.Fields(rest)...)
			}
		}
	}
	return patterns, nil
}

// checkContent returns what is wrong with an embedded file, or "". decode-go
// embeds its guest gzipped, so a .gz is inflated and its content checked
// too: a gzip of a pointer, or a truncated .gz, is as broken as a pointer.
// A .wasm, inflated or not, must start with the wasm magic.
func checkContent(target string, content []byte) string {
	if base, ok := strings.CutSuffix(target, ".gz"); ok {
		zr, err := gzip.NewReader(bytes.NewReader(content))
		if err != nil {
			return "which is not gzip"
		}
		inflated, err := io.ReadAll(zr)
		if err != nil {
			return "which does not inflate"
		}
		if bytes.HasPrefix(inflated, lfsPointerPrefix) {
			return "which inflates to a Git LFS pointer"
		}
		target, content = base, inflated
	}
	if strings.HasSuffix(target, ".wasm") && !bytes.HasPrefix(content, []byte("\x00asm")) {
		return "which is not a wasm module"
	}
	return ""
}
