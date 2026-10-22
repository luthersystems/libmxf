#!/usr/bin/env bash
# Copyright © 2021 Luther Systems, Ltd. All right reserved.

# This script runs source code checking tools it can be called via a make
# target or can be run automatically during CI testing.

set -exuo pipefail

golangci-lint run --timeout=5m -v
# scripts/modzipcheck is its own module, so the run above does not reach it.
(cd scripts/modzipcheck && golangci-lint run --timeout=5m -v)
