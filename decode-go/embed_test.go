package decode

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// guestSHA256 is the sha256 of the inflated guest, decode-wapc.wasm. A
// change to the embedded guest must change this line too, so a review sees
// it, and a corrupt or swapped .gz fails here. Rebuild with
// make -C decode-go/decode-wapc build, then set this to the sha256 it prints.
const guestSHA256 = "3116e55d802b34478b510c4f876cc346500d4a03d4dd1d1d0259d8e50defb509"

// TestEmbeddedGuestIsWasm checks that the embedded guest inflates to the
// pinned wasm module. The Go module zip holds what the git tree holds, so a
// guest tracked by Git LFS reaches every consumer as a 130-byte LFS pointer
// (libmxf#12). CI runs this test on a checkout without LFS, which sees the
// same bytes a consumer does. A local checkout with LFS content passes it
// either way; make modzip-check is the local check that reads what consumers
// get.
func TestEmbeddedGuestIsWasm(t *testing.T) {
	start := time.Now()
	wasm, err := guestWasm(guestDecodeWasmGz)
	t.Logf("inflated %d bytes to %d in %v", len(guestDecodeWasmGz), len(wasm), time.Since(start))
	require.NoError(t, err)
	sum := sha256.Sum256(wasm)
	require.Equal(t, guestSHA256, hex.EncodeToString(sum[:]))
}

// TestGuestWasmRefusesBadInput checks that a broken embed names its cause,
// instead of a gzip or wazero error.
func TestGuestWasmRefusesBadInput(t *testing.T) {
	pointer := []byte("version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:3116e55d802b34478b510c4f876cc346500d4a03d4dd1d1d0259d8e50defb509\nsize 8454492\n")
	_, err := guestWasm(pointer)
	require.ErrorContains(t, err, "Git LFS pointer")

	_, err = guestWasm([]byte("not gzip"))
	require.ErrorContains(t, err, "is not gzip")

	var notWasm bytes.Buffer
	zw := gzip.NewWriter(&notWasm)
	_, err = zw.Write([]byte("not wasm"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	_, err = guestWasm(notWasm.Bytes())
	require.ErrorContains(t, err, "does not hold a wasm module")

	_, err = guestWasm(notWasm.Bytes()[:notWasm.Len()-4])
	require.ErrorContains(t, err, "inflate")
}

// BenchmarkGuestWasm measures the one-time start-up cost of the gzip embed.
func BenchmarkGuestWasm(b *testing.B) {
	for b.Loop() {
		if _, err := guestWasm(guestDecodeWasmGz); err != nil {
			b.Fatal(err)
		}
	}
}
