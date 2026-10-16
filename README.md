# Luther MXF Library

This library provides off-chain tools to process message transformation (mxf) data that has been generated on-chain by the Luther platform.

See the [Luther Docs](https://docs.luthersystems.com/luther/platform/substrate) Private data section for a description of libmxf and private data processing.

## Private dependencies

The build for this package requires private dependencies from:
```
export GOPRIVATE="github.com/luthersystems/substrate"
```

## Rebuilding the decoder wasm

Both decoders embed a wasm build of the MXF codec, committed through Git LFS.
They build with the standard Go toolchain; TinyGo is not needed.

| Decoder | wasm | Rebuild |
|---|---|---|
| Go (`decode-go`) | `decode-go/decode-wapc/decode-wapc.wasm`, a waPC guest built as a WASI reactor (`GOOS=wasip1`, `-buildmode=c-shared`) | `make -C decode-go/decode-wapc build` |
| JavaScript (`decode-js/mxf`) | `decode-js/mxf/src/app.wasm` (`GOOS=js`) | `make -C decode-js app.wasm`, then copy `decode-js/app.wasm` to `decode-js/mxf/src/` |

`decode.Decode` compiles the guest once per process, on its first call, and
gives each call its own instance.
