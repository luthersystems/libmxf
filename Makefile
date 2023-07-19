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
go-citest: go-test
	@

.PHONY: js-citest
js-citest: js-test
	@

.PHONY: go-test
go-test:
	cd decode-js && make go-test
	cd decode-go && make test

.PHONY: static-checks
static-checks:
	export GOPRIVATE="github.com/luthersystems/substrate"
	./scripts/static-checks.sh

.PHONY: js-test
js-test:
	cd decode-js && make js-test
