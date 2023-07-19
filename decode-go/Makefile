GO_TEST_BASE=go test ${GO_TEST_FLAGS}
GO_TEST_TIMEOUT_10=${GO_TEST_BASE} -timeout 10m

.DEFAULT_GOAL := default
.PHONY: default
default: all

.PHONY: all
all:
	@

.PHONY: test
test:
	${GO_TEST_TIMEOUT_10} ./...

