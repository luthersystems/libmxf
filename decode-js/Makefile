APP_DEPS=$(shell find app)
WASM_OUT=app.wasm

WASM_GOOS=js
WASM_GOARCH=wasm
WASM_ENV=GOOS=${WASM_GOOS} GOARCH=${WASM_GOARCH}

GOROOT=$(shell go env GOROOT)
WASM_ROOT=${GOROOT}/misc/wasm

# WASM_RUNTIME is a javascript module that is required to run compiled Go.
WASM_RUNTIME=wasm_exec.js
WASM_RUNTIME_SOURCE=${WASM_ROOT}/${WASM_RUNTIME}

.PHONY: default
default: build

.PHONY: build
build: ${WASM_OUT}

.PHONY: clean
clean:
	$(MAKE) -C server clean
	rm -f ${WASM_RUNTIME} ${WASM_OUT}

.PHONY: run
run: ${WASM_RUNTIME} ${WASM_OUT}
	node ${WASM_RUNTIME} ${WASM_OUT}

.PHONY: go-test
go-test:
	${WASM_ENV} go test ./app/...
	go test ./server/...

.PHONY: serve
serve: ${WASM_OUT}
	$(MAKE) -C server
	./server/server

${WASM_RUNTIME}:
	cp ${WASM_RUNTIME_SOURCE} $@

${WASM_OUT}: ${APP_DEPS} ${WASM_RUNTIME}
	${WASM_ENV} go build -tags wasm -o $@ ./app

.PHONY: js-test
js-test:
	cd mxf && make test
