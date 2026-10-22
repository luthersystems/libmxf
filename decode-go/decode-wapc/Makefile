# The guest is embedded gzipped (decode-wapc.wasm.gz) so binaries that link
# decode-go are about 6 MB smaller. decode-wapc.wasm is an intermediate and is
# not committed.
#
# The build is reproducible: -trimpath drops local paths, -buildvcs=false
# drops the git commit, commit time and dirty flag that go build stamps into
# a binary by default, and -buildid= drops the build ID (a hash of the
# inputs, so even a comment edit changed it). go.mod's toolchain line is only
# a minimum, so GOTOOLCHAIN is set to it exactly, and the settings that change
# the code (GOWASM, GOEXPERIMENT, GOFLAGS) are cleared. So anyone can rebuild
# the wasm and get the sha256 pinned as guestSHA256 in
# decode-go/embed_test.go; make verify checks that, and CI runs it.
#
# gzip -n leaves out the name and time, so the same gzip gives the same .gz
# for the same wasm (another gzip version may not, which is why verify
# compares the wasm, not the .gz). The .gz is written to a temporary file and
# moved, so a failed gzip leaves the committed file as it was.

TOOLCHAIN=$(shell sed -n 's/^toolchain //p' ../../go.mod)
GO_BUILD=GOTOOLCHAIN=$(TOOLCHAIN) GOWASM= GOEXPERIMENT= GOFLAGS= GOOS=wasip1 GOARCH=wasm \
	go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags='-s -buildid='
PINNED_SHA256=$(shell sed -n 's/^const guestSHA256 = "\([0-9a-f]*\)"$$/\1/p' ../embed_test.go)
# SHA256 hashes stdin: sha256sum on Linux, shasum on macOS.
SHA256={ sha256sum 2>/dev/null || shasum -a 256; } | cut -d' ' -f1

.PHONY: build
build:
	@echo "----------"
	@echo "Building Go wasm Guest"
	@echo "----------"
	$(GO_BUILD) -o decode-wapc.wasm .
	gzip -9 -n -c decode-wapc.wasm > decode-wapc.wasm.gz.tmp
	mv decode-wapc.wasm.gz.tmp decode-wapc.wasm.gz
	@echo "sha256 $$(cat decode-wapc.wasm | $(SHA256)): set guestSHA256 in decode-go/embed_test.go to it"

# verify rebuilds the wasm from source and fails unless it matches both the
# pinned sha256 and the committed .gz, so the embedded guest is known to be
# this source and nothing else.
.PHONY: verify
verify:
	@test -n "$(TOOLCHAIN)" || { echo "no toolchain line in ../../go.mod"; exit 1; }
	@test -n "$(PINNED_SHA256)" || { echo "guestSHA256 not found in ../embed_test.go"; exit 1; }
	$(GO_BUILD) -o decode-wapc.wasm.verify .
	@got=$$(cat decode-wapc.wasm.verify | $(SHA256)); rm -f decode-wapc.wasm.verify; \
	embedded=$$(gzip -dc decode-wapc.wasm.gz | $(SHA256)); \
	echo "rebuilt with $(TOOLCHAIN): $$got"; echo "pinned guestSHA256: $(PINNED_SHA256)"; echo "committed .gz holds: $$embedded"; \
	ok=true; \
	if [ "$$got" != "$(PINNED_SHA256)" ]; then ok=false; \
		echo "FAIL: this source does not build the pinned guest; run make build, commit the .gz and set guestSHA256"; fi; \
	if [ "$$embedded" != "$(PINNED_SHA256)" ]; then ok=false; \
		echo "FAIL: the committed decode-wapc.wasm.gz is not the pinned guest"; fi; \
	$$ok && echo "ok: the embedded guest is built from this source"
