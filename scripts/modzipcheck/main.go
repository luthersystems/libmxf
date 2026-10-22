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
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"
)

const modulePath = "github.com/luthersystems/libmxf"

var lfsPointerPrefix = []byte("version https://git-lfs.github.com/spec/")

// embedRE matches a //go:embed line. Patterns are split on spaces; quoted
// patterns and globs are not used in libmxf, and a pattern this does not
// resolve fails the check, so a new form cannot pass unchecked.
var embedRE = regexp.MustCompile(`(?m)^//go:embed\s+(.+)$`)

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
		for _, match := range embedRE.FindAllSubmatch(files[name], -1) {
			for _, pattern := range strings.Fields(string(match[1])) {
				target := path.Join(path.Dir(name), pattern)
				embedded[target] = true
				content, ok := files[target]
				switch {
				case !ok:
					bad = append(bad, fmt.Sprintf("%s embeds %s, which is not a file in the zip", name, target))
				case bytes.HasPrefix(content, lfsPointerPrefix):
					bad = append(bad, fmt.Sprintf("%s embeds %s, which is a Git LFS pointer in the zip", name, target))
				default:
					fmt.Printf("ok: %s embeds %s (%d bytes)\n", name, target, len(content))
				}
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
