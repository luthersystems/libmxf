# The guest is embedded gzipped (decode-wapc.wasm.gz) so binaries that link
# decode-go are about 6 MB smaller. decode-wapc.wasm is an intermediate and is
# not committed. gzip -n leaves out the name and time, so the same gzip gives
# the same .gz for the same wasm (another gzip version may not). The .gz is
# written to a temporary file and moved, so a failed gzip leaves the committed
# file as it was. Set guestSHA256 in decode-go/embed_test.go to the sha256
# this prints.
build:
	@echo "----------"
	@echo "Building Go wasm Guest"
	@echo "----------"
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -ldflags=-s -o decode-wapc.wasm .
	gzip -9 -n -c decode-wapc.wasm > decode-wapc.wasm.gz.tmp
	mv decode-wapc.wasm.gz.tmp decode-wapc.wasm.gz
	@# sha256sum on Linux, shasum on macOS.
	sha256sum decode-wapc.wasm 2>/dev/null || shasum -a 256 decode-wapc.wasm
