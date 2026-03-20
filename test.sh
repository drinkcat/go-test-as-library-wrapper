#!/bin/bash

set -e -x

go build

for dir in test/example*/; do
    rm -f "$dir/test_as_library_entrypoint.go"
    ./go-test-as-library-wrapper "$dir"
    test/build-and-test.sh "$dir"
done
