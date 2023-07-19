package main

import (
	"github.com/luthersystems/elps/lisp"
	"github.com/luthersystems/libmxf"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"decode": decode,
	})
}

func decode(enc []byte) ([]byte, error) {
	getKey := func(dsid string) string {
		key, err := wapc.HostCall("", "libmxf", "get_key", []byte(dsid))
		if err != nil {
			return ""
		}
		return string(key)
	}

	respVal := libmxf.MxfDecode(getKey, string(enc))

	if len(respVal.Bytes()) == 0 {
		return nil, lisp.GoError(respVal)
	}

	return respVal.Bytes(), nil
}
