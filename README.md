# go-test-as-library-wrapper

A workaround tool that enables Go tests to be built as C libraries (`c-archive` or `c-shared`).

## Rationale

On platforms like iOS and Android, running standalone Go test binaries isn't feasible. The natural solution would be `go test -buildmode=c-archive`, but the resulting library cannot actually be used for testing — the `main` function is never called, and the test runner is not initialized. See [golang/go#77524](https://github.com/golang/go/issues/77524) for the upstream discussion.

`go-test-as-library-wrapper` is a stopgap until this is supported natively by the Go toolchain (if ever).

## How it works

Run the tool against a package directory to generate `test_as_library_entrypoint.go`. This file exports a C-callable function takes arguments, just
like a usual go test binary:

```c
int go_test_main_with_args(int argc, const char** argv);
```

The generated file uses the `//go:build test_archive` constraint, so it only activates when you explicitly opt in.

## Usage

```bash
# Install
go install github.com/drinkcat/go-test-as-library-wrapper@latest

# Generate the entrypoint for your package
go-test-as-library-wrapper ./path/to/package

# Build as static library
CGO_ENABLED=1 go test -tags test_archive -buildmode=c-archive -c -o test.a

# Build as shared library
CGO_ENABLED=1 go test -tags test_archive -buildmode=c-shared -c -o test.so
```

Then call from C (this enables verbose output):

```c
#include <stdio.h>

int go_test_main_with_args(int argc, const char **argv);

int main() {
	int ret = go_test_main_with_args(2, (const char*[]){"<none>", "-test.v"});
	fprintf(stderr,"Go test main returned: %d\n", ret);
	return ret;
}
```

See [example/](example/) for a complete working example including static and dynamic linking.

## Limitations

- `go_test_main_with_args` can only be called once.
- Benchmarks, fuzz targets, and examples are not supported — only tests.
- Build tags (`-tags`) are not taken into account when scanning for tests, so files with `//go:build` constraints may be incorrectly included.
- This relies on unofficial testing API, and may break at any time,
but this is tested with Go 1.26.
