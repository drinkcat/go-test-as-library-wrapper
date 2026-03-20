#!/bin/bash

set -e -x

DIR="${1:?Usage: build-and-test.sh <dir>}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$DIR"

# Build and run normal executable
CGO_ENABLED=1 go test -c -o test
./test -test.v | tee test.log

# Build and run with static linkage
CGO_ENABLED=1 go test -tags test_archive -buildmode=c-archive -c -o test-static.a
gcc -c -o ctest-static.o "$SCRIPT_DIR/c-harness/ctest-static.c"
gcc -o ctest-static ctest-static.o test-static.a

# Pass `--help` to test that args are overriden
./ctest-static --help | tee ctest-static.log

# Build and run with dynamic linkage
CGO_ENABLED=1 go test -tags test_archive -buildmode=c-shared -c -o test-dynamic.so
gcc -o ctest-dynamic "$SCRIPT_DIR/c-harness/ctest-dynamic.c" -ldl

./ctest-dynamic --help | tee ctest-dynamic.log

diff test.log ctest-static.log
diff test.log ctest-dynamic.log
