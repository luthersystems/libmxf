# Copyright © 2023 Luther Systems, Ltd. All right reserved.

# Makefile
#
# The primary project makefile that should be run from the root directory and is
# able to build and run the entire application.

.DEFAULT_GOAL := default
.PHONY: default
default: all

.PHONY: all
all:
	@

.PHONY: go-citest
go-citest: static-checks go-test
	@

GO_TEST_BASE=go test ${GO_TEST_FLAGS}
GO_TEST_TIMEOUT_10=${GO_TEST_BASE} -timeout 10m

.PHONY: go-test
go-test:
	export GOPRIVATE="github.com/luthersystems/substrate"
	${GO_TEST_TIMEOUT_10} ./...

.PHONY: static-checks
static-checks:
	export GOPRIVATE="github.com/luthersystems/substrate"
	./scripts/static-checks.sh
