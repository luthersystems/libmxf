# Luther MXF Library

This library provides off-chain tools to process message transformation (mxf) data that has been generated on-chain by the Luther platform.

See the [Luther Docs](https://docs.luthersystems.com/luther/platform/substrate) Private data section for a description of libmxf and private data processing.

## Private dependencies

The build for this package requires private dependencies from:
```
export GOPRIVATE="github.com/luthersystems/substrate"
```

## Rebuilding the decoder wasm

Both decoders embed a wasm build of the MXF codec. They build with the
standard Go toolchain; TinyGo is not needed.

The Go guest is committed gzipped, as `decode-wapc.wasm.gz`, so every binary
that links `decode-go` is about 6 MB smaller (2.2 MB in place of 8.4 MB).
`decode.Decode` inflates it once, before its first compile; that takes about
0.1 s, next to a compile of about 5 s. A rebuild must also update the pinned
sha256 of the inflated wasm, `guestSHA256` in `decode-go/embed_test.go`.
The build is reproducible (no paths, git stamp or build ID in the wasm, and
go.mod pins the toolchain), so `make -C decode-go/decode-wapc verify`, run
in CI, checks that the committed guest is built from this source.

It is a normal git file, not a Git LFS file.
`go get` and the Go module proxy build the module from the git tree and never
fetch LFS content, so an LFS-tracked guest reaches every consumer as a
130-byte pointer (libmxf#12). Do not track it with LFS again; the Go CI job
checks out without LFS, and `make modzip-check` reads the module zip, to catch that. The JavaScript `app.wasm` stays in LFS,
because npm publishes it from a full checkout and no Go code embeds it.

| Decoder | wasm | Rebuild |
|---|---|---|
| Go (`decode-go`) | `decode-go/decode-wapc/decode-wapc.wasm.gz`, a gzip of a waPC guest built as a WASI reactor (`GOOS=wasip1`, `-buildmode=c-shared`) | `make -C decode-go/decode-wapc build` |
| JavaScript (`decode-js/mxf`) | `decode-js/mxf/src/app.wasm` (`GOOS=js`) | `make -C decode-js app.wasm`, then copy `decode-js/app.wasm` and `decode-js/wasm_exec.js` to `decode-js/mxf/src/` |

`wasm_exec.js` is the JavaScript runtime of the Go toolchain that built
`app.wasm`, and it must come from that same Go release. The Makefile copies it
from `$(go env GOROOT)/lib/wasm`.

`decode.Decode` compiles the guest once per process, on its first call, and
gives each call its own instance.
