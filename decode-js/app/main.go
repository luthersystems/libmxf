//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/luthersystems/elps/lisp"
	"github.com/luthersystems/libmxf"
)

func JSMxfDecode(this js.Value, vals []js.Value) interface{} {
	if len(vals) != 2 {
		libmxf.Log(fmt.Sprintf("invalid args: [%d] != 2", len(vals)))
		return nil
	}
	encMsg := vals[0].String()
	callback := vals[1]

	lval := libmxf.MxfDecode(func(dsid string) string {
		return callback.Call("getKey", js.ValueOf(dsid)).String()
	}, encMsg)
	err := lisp.GoError(lval)
	if err != nil {
		callback.Call("reject", js.ValueOf(err.Error()))
	} else {
		callback.Call("resolve", js.ValueOf(string(lval.Bytes())))
	}
	return nil
}

func main() {
	// hook up "exported" functions

	done := make(chan struct{})
	var exportCallbacks []js.Func
	export := func(name string, fn func(js.Value, []js.Value) interface{}) js.Func {
		cb := js.FuncOf(fn)
		js.Global().Set(name, cb)
		exportCallbacks = append(exportCallbacks, cb)
		return cb
	}
	export("MxfDecode", JSMxfDecode)
	export("GoKill", func(js.Value, []js.Value) interface{} {
		close(done)
		return nil
	})

	// initialization complete

	libmxf.Log("loaded wasm")

	<-done
	for i := range exportCallbacks {
		exportCallbacks[i].Release()
	}
	exportCallbacks = nil
}
