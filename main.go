// go-test-archive-wrapper generates a test_archive_entrypoint.go file for a Go package,
// enabling the package's tests to be built as a C archive or shared library.
//
// Usage: go-test-archive-wrapper [directory]
//
// If directory is omitted, the current directory is used.
// The generated file is written to test_archive_entrypoint.go in the target directory.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const entrypointTemplate = `// This generated file can be used as a C entry point to run Go tests in a c-archive or c-shared
// library.
// A more proper solution would require a Go toolchain change, see
// https://github.com/golang/go/issues/77524 .
//
// To use this code, build tests as an archive with
// CGO_ENABLED=1 go test -tags test_archive -buildmode=c-archive -c -o test-static.a
// or
// CGO_ENABLED=1 go test -tags test_archive -buildmode=c-shared -c -o test-dynamic.so
// and then link the resulting library into your C code, and call the ` + "`go_test_main_with_args`" + `
// with this signature:
// extern int go_test_main_with_args(int argc, char** argv);
//
// This file is only built when then test_archive tag is set, so it will not be built
// in the normal app or native test.

//go:build test_archive

package {{.Package}}

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

// HACK: This is an unfortunate copy-paste of code in the Go standard library's testing package.
// We need this as we hijack the ` + "`MainStart`" + ` function that is not really meant to
// be called from user code, and it expects a ` + "`testing.testDeps`" + ` implementation to be passed in,
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
{{- range .Tests}}
			{Name: "{{.}}", F: {{.}}},
{{- end}}
		},
		nil, // TODO: support benchmarks
		nil, // TODO: support fuzz targets
		nil, // TODO: support examples
	)
{{- if .HasTestMain}}
	TestMain(m)
	return C.int(reflect.ValueOf(m).Elem().FieldByName("exitCode").Int())
{{- else}}
	return C.int(m.Run())
{{- end}}
}
`

type templateData struct {
	Package     string
	Tests       []string
	HasTestMain bool
}

// TODO: Support build tags (-tags) so that files with //go:build constraints are
// correctly included or excluded, matching what `go test -tags ...` would see.
func findTests(dir string) (pkg string, tests []string, hasTestMain bool, err error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return "", nil, false, fmt.Errorf("parsing %s: %w", dir, err)
	}

	for pkgName, pkgAst := range pkgs {
		// Skip external test packages (foo_test)
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}
		pkg = pkgName
		for _, file := range pkgAst.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil {
					continue
				}
				name := fn.Name.Name
				if name == "TestMain" {
					hasTestMain = true
				} else if isTestFunc(name, fn) {
					tests = append(tests, name)
				}
			}
		}
	}

	if pkg == "" {
		return "", nil, false, fmt.Errorf("no non-external test package found in %s", dir)
	}
	return pkg, tests, hasTestMain, nil
}

// isTestFunc checks if a function is a Test function (TestXxx with *testing.T parameter).
func isTestFunc(name string, fn *ast.FuncDecl) bool {
	if !strings.HasPrefix(name, "Test") || len(name) == 4 {
		return false
	}
	// Must have exactly one parameter of type *testing.T
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	param := fn.Type.Params.List[0]
	star, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "T"
}

func run() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [directory]\n\nGenerates test_archive_entrypoint.go in the given directory (default: current directory).\n", os.Args[0])
	}
	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	pkg, tests, hasTestMain, err := findTests(absDir)
	if err != nil {
		return err
	}

	data := templateData{
		Package:     pkg,
		Tests:       tests,
		HasTestMain: hasTestMain,
	}

	tmpl, err := template.New("entrypoint").Parse(entrypointTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	outPath := filepath.Join(absDir, "test_archive_entrypoint.go")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	fmt.Printf("Generated %s (package %s, %d tests, TestMain=%v)\n", outPath, pkg, len(tests), hasTestMain)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
