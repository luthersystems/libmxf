// Command decode-wapc is the waPC guest that decodes MXF messages. Build it
// with the standard Go toolchain as a WASI reactor:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o decode-wapc.wasm .
//
// make build does that and gzips the result to decode-wapc.wasm.gz, the file
// decode-go embeds.
//
// The host instantiates WASI, calls _initialize (which runs package init),
// then calls __guest_call for each request.
package main

import (
	"github.com/luthersystems/elps/lisp"
	"github.com/luthersystems/libmxf"
)

func main() {}

type guestFunction func(payload []byte) ([]byte, error)

// functions maps a waPC operation name to its handler.
var functions = map[string]guestFunction{}

func registerFunction(name string, fn guestFunction) { functions[name] = fn }

func init() {
	registerFunction("decode", decode)
}

func decode(enc []byte) ([]byte, error) {
	getKey := func(dsid string) string {
		key, err := hostCall("", "libmxf", "get_key", []byte(dsid))
		if err != nil {
			return ""
		}
		return string(key)
	}

	respVal := libmxf.MxfDecode(getKey, string(enc))

	if respVal.Type != lisp.LBytes || len(respVal.Bytes()) == 0 {
		return nil, lisp.GoError(respVal)
	}

	return respVal.Bytes(), nil
}
