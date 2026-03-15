#!/bin/bash

set -e -x

go build

cd example
rm -f test_as_library_entrypoint.go
../go-test-as-library-wrapper .
./build-and-test.sh
