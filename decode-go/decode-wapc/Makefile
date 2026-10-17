build:
	@echo "----------"
	@echo "Building Go wasm Guest"
	@echo "----------"
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -ldflags=-s -o decode-wapc.wasm .
