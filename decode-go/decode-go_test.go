package decode

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	getKey := func(dsid string) string {
		keys := map[string]string{
			"89871b002200cbd38dec6cc43988e687f0a6799e2734635fee7cc891eca1170d-1": "22c20ac7e93194e19661cb55fc3633bbfd1a44c82d852fc8ee5e160aaba7c9a0",
		}
		return keys[dsid]
	}

	encMsg := map[string]interface{}{
		"message": nil,
		"mxf":     "v1",
		"transforms": []map[string]interface{}{
			{
				"body": map[string]string{
					"dsid":             "89871b002200cbd38dec6cc43988e687f0a6799e2734635fee7cc891eca1170d-1",
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

	encMsgJSON := string(b)

	decodedMsg, err := Decode(context.Background(), getKey, encMsgJSON)
	require.NoError(t, err, "Decode failed")

	expected := `{"client_id":"dba7bd5f-7a8f-4797-9851-202885837845","client_originator":"cbfc1ad7-84ec-4a4b-8302-7b869c00bd8f","field_mask":["client_originator"]}`

	assert.Equal(t, expected, decodedMsg, "Decoded message did not match expected")
}
