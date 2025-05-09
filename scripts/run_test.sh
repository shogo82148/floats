#!/usr/bin/env bash

set -eux

SEED=${GITHUB_RUN_ID:-$(date +%s)}
echo "$SEED"

# shellcheck disable=SC2206
TEST_NAMES=(${1//,/ })

ROOT=$(cd "$(dirname "$0")"; cd ..; pwd)
cd "$ROOT"

for TEST_NAME in "${TEST_NAMES[@]}"; do
  if [[ $TEST_NAME =~ _to_[iu]64$ ]]; then
    "$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" -rminMag "$TEST_NAME" | go run ./internal/cmd/float_test "$TEST_NAME"
  elif [[ $TEST_NAME =~ ^f(16|32|64|128)_to_f(16|32|64|128)$ ]]; then
    "$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" "$TEST_NAME" | go run ./internal/cmd/float_test "$TEST_NAME"
  else
    "$ROOT/bin/testfloat_gen" -seed "$SEED" "$TEST_NAME" | go run ./internal/cmd/float_test "$TEST_NAME"
  fi
done
