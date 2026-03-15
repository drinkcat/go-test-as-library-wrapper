#!/bin/bash

set -e -x

go build

cd example
../go-test-as-library-wrapper .
./build-and-test.sh
