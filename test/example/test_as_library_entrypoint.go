// This generated file can be used as a C entry point to run Go tests in a c-archive or c-shared
// library.
// A more proper solution would require a Go toolchain change, see
// https://github.com/golang/go/issues/77524 .
//
// To regenerate this file, run: go run github.com/drinkcat/go-test-as-library-wrapper@latest .
//
// To use this code, build tests as an archive with
// CGO_ENABLED=1 go test -tags test_archive -buildmode=c-archive -c -o test-static.a
// or
// CGO_ENABLED=1 go test -tags test_archive -buildmode=c-shared -c -o test-dynamic.so
// and then link the resulting library into your C code, and call the `go_test_main_with_args`
// with this signature:
// extern int go_test_main_with_args(int argc, char** argv);
//
// This file is only built when then test_archive tag is set, so it will not be built
// in the normal app or native test.

//go:build test_archive

package main

import "C"

import (
	"io"
	"os"
	"reflect"
	"regexp"
	"runtime/pprof"
	"testing"
	"time"
	"unsafe"
)

// HACK: This is an unfortunate stubbing of code in the Go standard library's testing package.
// We need this as we hijack the `MainStart` function that is not really meant to
// be called from user code, and it expects a `testing.testDeps` implementation to be passed in,
// which is internal, and cannot be imported.
type corpusEntry = struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []any
	Generation int
	IsSeed     bool
}

type minimalTestDeps struct{}

var matchPat string
var matchRe *regexp.Regexp

func (minimalTestDeps) MatchString(pat, str string) (bool, error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		var err error
		matchRe, err = regexp.Compile(pat)
		if err != nil {
			return false, err
		}
	}
	return matchRe.MatchString(str), nil
}
func (minimalTestDeps) ImportPath() string                        { return "" }
func (minimalTestDeps) ModulePath() string                        { return "" }
func (minimalTestDeps) SetPanicOnExit0(bool)                      {}
func (minimalTestDeps) StartCPUProfile(w io.Writer) error         { return pprof.StartCPUProfile(w) }
func (minimalTestDeps) StopCPUProfile()                           { pprof.StopCPUProfile() }
func (minimalTestDeps) StartTestLog(io.Writer)                    {}
func (minimalTestDeps) StopTestLog() error                        { return nil }
func (minimalTestDeps) WriteProfileTo(name string, w io.Writer, debug int) error {
	return pprof.Lookup(name).WriteTo(w, debug)
}
func (minimalTestDeps) CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []corpusEntry, []reflect.Type, string, string) error {
	return nil
}
func (minimalTestDeps) RunFuzzWorker(func(corpusEntry) error) error             { return nil }
func (minimalTestDeps) ReadCorpus(string, []reflect.Type) ([]corpusEntry, error) { return nil, nil }
func (minimalTestDeps) CheckCorpus([]any, []reflect.Type) error                 { return nil }
func (minimalTestDeps) ResetCoverage()                                          {}
func (minimalTestDeps) SnapshotCoverage()                                       {}
func (minimalTestDeps) InitRuntimeCoverage() (string, func(string, string) (string, error), func() float64) {
	return "", nil, nil
}
// End of unfortunate hack.

//export go_test_main_with_args
func go_test_main_with_args(argc C.int, argv **C.char) C.int {
	os.Args = make([]string, int(argc))
	for i := range os.Args {
		os.Args[i] = C.GoString(*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(argv)) + uintptr(i)*unsafe.Sizeof(*argv))))
	}

	m := testing.MainStart(
		minimalTestDeps{},
		[]testing.InternalTest{
			{Name: "TestAdd", F: TestAdd},
		},
		nil, // TODO: support benchmarks
		nil, // TODO: support fuzz targets
		nil, // TODO: support examples
	)

	// Note: This looks somewhat terrible, but replicates what `go test` does.
	// (see output of `go test -work -c -o test`)
	TestMain(m)
	return C.int(reflect.ValueOf(m).Elem().FieldByName("exitCode").Int())
}
