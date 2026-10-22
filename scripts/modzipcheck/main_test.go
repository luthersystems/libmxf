package main

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func gz(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestCheckContent checks each way an embedded guest can reach consumers
// broken, gzipped or not.
func TestCheckContent(t *testing.T) {
	wasm := []byte("\x00asm\x01\x00\x00\x00")
	pointer := []byte("version https://git-lfs.github.com/spec/v1\noid sha256:00\nsize 1\n")
	good := gz(t, wasm)
	for _, tc := range []struct {
		name, target string
		content      []byte
		want         string
	}{
		{"wasm", "g.wasm", wasm, ""},
		{"gzipped wasm", "g.wasm.gz", good, ""},
		{"other file", "notes.txt", []byte("hello"), ""},
		{"not wasm", "g.wasm", []byte("hello"), "which is not a wasm module"},
		{"not gzip", "g.wasm.gz", []byte("hello"), "which is not gzip"},
		{"truncated gzip", "g.wasm.gz", good[:len(good)-4], "which does not inflate"},
		{"gzipped pointer", "g.wasm.gz", gz(t, pointer), "which inflates to a Git LFS pointer"},
		{"gzipped non-wasm", "g.wasm.gz", gz(t, []byte("hello")), "which is not a wasm module"},
	} {
		if got := checkContent(tc.target, tc.content); got != tc.want {
			t.Errorf("%s: checkContent = %q, want %q", tc.name, got, tc.want)
		}
	}
}
