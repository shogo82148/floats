#!/usr/bin/env bash

set -eux

SEED=${GITHUB_RUN_ID:-$(date +%s)}
echo "$SEED"

ROOT=$(cd "$(dirname "$0")"; cd ..; pwd)
cd "$ROOT"

### convert float to float tests

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f16_to_f32 | go run ./internal/cmd/float_test f16_to_f32
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f16_to_f64 | go run ./internal/cmd/float_test f16_to_f64
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f16_to_f128 | go run ./internal/cmd/float_test f16_to_f128

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f32_to_f16 | go run ./internal/cmd/float_test f32_to_f16
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f32_to_f64 | go run ./internal/cmd/float_test f32_to_f64
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f32_to_f128 | go run ./internal/cmd/float_test f32_to_f128

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f64_to_f16 | go run ./internal/cmd/float_test f64_to_f16
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f64_to_f32 | go run ./internal/cmd/float_test f64_to_f32
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f64_to_f128 | go run ./internal/cmd/float_test f64_to_f128

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f128_to_f16 | go run ./internal/cmd/float_test f128_to_f16
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f128_to_f32 | go run ./internal/cmd/float_test f128_to_f32
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f128_to_f64 | go run ./internal/cmd/float_test f128_to_f64


### convert float to int tests

"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" -rminMag f16_to_i64 | go run ./internal/cmd/float_test f16_to_i64
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" -rminMag f32_to_i64 | go run ./internal/cmd/float_test f32_to_i64
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" -rminMag f64_to_i64 | go run ./internal/cmd/float_test f64_to_i64
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" -rminMag f128_to_i64 | go run ./internal/cmd/float_test f128_to_i64

### f16 operations
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f16_mul | go run ./internal/cmd/float_test f16_mul

### f32 operations
"$ROOT/bin/testfloat_gen" -level 2 -seed "$SEED" f32_mul | go run ./internal/cmd/float_test f32_mul
