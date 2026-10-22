package decode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEmbeddedGuestIsWasm checks that the embedded guest is a wasm module.
// The Go module zip holds what the git tree holds, so a guest tracked by Git
// LFS reaches every consumer as a 130-byte LFS pointer (libmxf#12). CI runs
// this test on a checkout without LFS, which sees the same bytes a consumer
// does.
func TestEmbeddedGuestIsWasm(t *testing.T) {
	require.True(t, bytes.HasPrefix(guestDecodeWasm, wasmMagic),
		"decode-wapc.wasm does not start with the wasm magic; first bytes: %q", head(guestDecodeWasm))
	require.NoError(t, checkGuest(guestDecodeWasm))
}

// TestCheckGuestRefusesLFSPointer checks that Decode names the cause when
// the embedded guest is an LFS pointer, instead of a wazero compile error.
func TestCheckGuestRefusesLFSPointer(t *testing.T) {
	pointer := []byte("version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:3116e55d802b34478b510c4f876cc346500d4a03d4dd1d1d0259d8e50defb509\nsize 8454492\n")
	err := checkGuest(pointer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Git LFS pointer")

	err = checkGuest([]byte("not wasm"))
	require.Error(t, err)
	require.NotContains(t, err.Error(), "Git LFS pointer")
}
