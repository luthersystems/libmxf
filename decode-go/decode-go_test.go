package decode

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDSID = "89871b002200cbd38dec6cc43988e687f0a6799e2734635fee7cc891eca1170d-1"
	testKey  = "22c20ac7e93194e19661cb55fc3633bbfd1a44c82d852fc8ee5e160aaba7c9a0"
	expected = `{"client_id":"dba7bd5f-7a8f-4797-9851-202885837845","client_originator":"cbfc1ad7-84ec-4a4b-8302-7b869c00bd8f","field_mask":["client_originator"]}`
)

func getTestKey(dsid string) string {
	keys := map[string]string{testDSID: testKey}
	return keys[dsid]
}

func testMessage(t testing.TB) string {
	encMsg := map[string]interface{}{
		"message": nil,
		"mxf":     "v1",
		"transforms": []map[string]interface{}{
			{
				"body": map[string]string{
					"dsid":             testDSID,
					"encrypted_base64": "Y4vQZSIr88swK2GTTZR65v7jh/q6Z9VtITJ1MAcShAg5oqvL5dHBz8Iy4u+I7D8rHJHTiGY9MyQnJGQC6OE6WPkA0UQJAuEIyWnrnBCagMBnNG0qFfcmQ2/GdY7sI3oWTATA5La2KkYd662y2n1Jig9R7QKeOo8d0ANnKqyn2qsfBOXV6qnvAWvbOTOsxnMYrZ/VpQC6F+w=",
				},
				"context_path": ".",
				"header": map[string]interface{}{
					"compressor":    "zlib",
					"encryptor":     "AES-256",
					"private_paths": []string{"."},
					"profile_paths": []string{".client_id"},
				},
			},
		},
	}

	b, err := json.Marshal(encMsg)
	require.NoError(t, err, "Failed to marshal encMsg")
	return string(b)
}

func TestDecode(t *testing.T) {
	decodedMsg, err := Decode(context.Background(), getTestKey, testMessage(t))
	require.NoError(t, err, "Decode failed")
	assert.Equal(t, expected, decodedMsg, "Decoded message did not match expected")
}

// TestDecodeCompilesOnce checks that the guest is compiled once per process:
// a second Decode reuses the compiled module. It checks the module's
// identity, not timing: an instantiation alone takes about 0.85 s under
// -race, so a time bound would fail on a slow runner.
func TestDecodeCompilesOnce(t *testing.T) {
	enc := testMessage(t)
	ctx := context.Background()

	start := time.Now()
	_, err := Decode(ctx, getTestKey, enc)
	require.NoError(t, err)
	first := time.Since(start)
	module, err := compiledGuest(context.Background())
	require.NoError(t, err)

	start = time.Now()
	decodedMsg, err := Decode(ctx, getTestKey, enc)
	require.NoError(t, err)
	second := time.Since(start)
	assert.Equal(t, expected, decodedMsg)

	again, err := compiledGuest(context.Background())
	require.NoError(t, err)
	assert.Same(t, module, again, "Decode compiled the guest again")
	t.Logf("first Decode %v (includes the compile unless an earlier test ran), second Decode %v", first, second)
}

// TestDecodeConcurrent checks that concurrent calls share the compiled guest
// and that each call's host calls reach its own getKey.
func TestDecodeConcurrent(t *testing.T) {
	enc := testMessage(t)

	const n = 8
	var calls [n]atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			getKey := func(dsid string) string {
				calls[i].Add(1)
				return getTestKey(dsid)
			}
			decodedMsg, err := Decode(context.Background(), getKey, enc)
			assert.NoError(t, err)
			assert.Equal(t, expected, decodedMsg)
		}(i)
	}
	wg.Wait()
	for i := range calls {
		assert.Equal(t, int32(1), calls[i].Load(), "call %d: getKey calls", i)
	}
}

func BenchmarkDecode(b *testing.B) {
	enc := testMessage(b)
	ctx := context.Background()
	_, err := Decode(ctx, getTestKey, enc) // compile outside the timed loop
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Decode(ctx, getTestKey, enc); err != nil {
			b.Fatal(err)
		}
	}
}

// TestDecodeUnknownKey checks that a message whose key the host does not
// know fails with an error, not a guest panic.
func TestDecodeUnknownKey(t *testing.T) {
	noKey := func(string) string { return "" }
	_, err := Decode(context.Background(), noKey, testMessage(t))
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unreachable", "the guest panicked")
}

// TestDecodeLeavesCompileWaitWhenContextEnds holds the compile lock, as a
// first compile does, and checks that a caller whose deadline passes while it
// waits returns the context's error instead of waiting for the lock.
func TestDecodeLeavesCompileWaitWhenContextEnds(t *testing.T) {
	guestLock <- struct{}{}
	defer func() { <-guestLock }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := Decode(ctx, getTestKey, testMessage(t))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second)
}

// TestDecodeContextAlreadyDone checks that a caller whose context has ended
// gets the context's error without waiting on the guest.
func TestDecodeContextAlreadyDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Decode(ctx, getTestKey, testMessage(t))
	assert.ErrorIs(t, err, context.Canceled)
}
