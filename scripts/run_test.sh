#!/usr/bin/env bash

set -eux

SEED=${GITHUB_RUN_ID:-$(date +%s)}
echo "$SEED"

TEST_NAME=$1
ROOT=$(cd "$(dirname "$0")"; cd ..; pwd)
cd "$ROOT"

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" "$TEST_NAME" | go run ./internal/cmd/float_test "$TEST_NAME"
