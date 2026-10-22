# The guest is embedded gzipped (decode-wapc.wasm.gz) so binaries that link
# decode-go are about 6 MB smaller. decode-wapc.wasm is an intermediate and is
# not committed.
#
# The build is reproducible: -trimpath drops local paths, -buildvcs=false
# drops the git commit, commit time and dirty flag that go build stamps into
# a binary by default, -buildid= drops the build ID (a hash of the inputs,
# so even a comment edit changed it), and go.mod's toolchain line fixes the
# compiler. So
# anyone can rebuild the wasm and get the sha256 pinned as guestSHA256 in
# decode-go/embed_test.go; make verify checks that, and CI runs it.
#
# gzip -n leaves out the name and time, so the same gzip gives the same .gz
# for the same wasm (another gzip version may not, which is why verify
# compares the wasm, not the .gz). The .gz is written to a temporary file and
# moved, so a failed gzip leaves the committed file as it was.

GO_BUILD=GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags='-s -buildid='
PINNED_SHA256=$(shell sed -n 's/^const guestSHA256 = "\([0-9a-f]*\)"$$/\1/p' ../embed_test.go)

.PHONY: build
build:
	@echo "----------"
	@echo "Building Go wasm Guest"
	@echo "----------"
	$(GO_BUILD) -o decode-wapc.wasm .
	gzip -9 -n -c decode-wapc.wasm > decode-wapc.wasm.gz.tmp
	mv decode-wapc.wasm.gz.tmp decode-wapc.wasm.gz
	@# sha256sum on Linux, shasum on macOS.
	sha256sum decode-wapc.wasm 2>/dev/null || shasum -a 256 decode-wapc.wasm

# verify rebuilds the wasm from source and fails unless it matches both the
# pinned sha256 and the committed .gz, so the embedded guest is known to be
# this source and nothing else.
.PHONY: verify
verify:
	@test -n "$(PINNED_SHA256)" || { echo "guestSHA256 not found in ../embed_test.go"; exit 1; }
	$(GO_BUILD) -o decode-wapc.wasm.verify .
	@got=$$( (sha256sum decode-wapc.wasm.verify 2>/dev/null || shasum -a 256 decode-wapc.wasm.verify) | cut -d' ' -f1); \
	embedded=$$(gzip -dc decode-wapc.wasm.gz | (sha256sum 2>/dev/null || shasum -a 256) | cut -d' ' -f1); \
	rm -f decode-wapc.wasm.verify; \
	echo "rebuilt  $$got"; echo "pinned   $(PINNED_SHA256)"; echo "embedded $$embedded"; \
	test "$$got" = "$(PINNED_SHA256)" && test "$$embedded" = "$(PINNED_SHA256)"
