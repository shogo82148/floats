#!/usr/bin/env bash

set -euxo pipefail

ROOT=$(cd "$(dirname "$0")"; cd ..; pwd)
rm -rf "$ROOT/test_float"
mkdir -p "$ROOT/test_float"
cd "$ROOT/test_float"

# download Berkeley SoftFloat and TestFloat
# http://www.jhauser.us/arithmetic/SoftFloat.html
curl -sSL -o SoftFloat-3e.zip http://www.jhauser.us/arithmetic/SoftFloat-3e.zip
curl -sSL -o TestFloat-3e.zip http://www.jhauser.us/arithmetic/TestFloat-3e.zip

if command -v gsha256sum >/dev/null 2>&1; then
    # gsha256sum is available on macOS
    HASH_CMD="gsha256sum"
else
    HASH_CMD="sha256sum"
fi

"$HASH_CMD" --check <<__SHA256__
21130ce885d35c1fe73fc1e1bf2244178167e05c6747cad5f450cc991714c746  SoftFloat-3e.zip
6d4bdf0096b48a653aa59fc203a9e5fe18b5a58d7a1b715107c7146776a0aad6  TestFloat-3e.zip
__SHA256__

JOBS=$(nproc 2>/dev/null || echo 1)

# build SoftFloat
unzip SoftFloat-3e.zip
pushd SoftFloat-3e/build/Linux-x86_64-GCC
make -j"$JOBS" all
popd

# build TestFloat
unzip TestFloat-3e.zip
pushd TestFloat-3e/build/Linux-x86_64-GCC
make -j"$JOBS" all
popd

mkdir -p "$ROOT/bin"
cp TestFloat-3e/build/Linux-x86_64-GCC/testfloat_gen "$ROOT/bin/"
