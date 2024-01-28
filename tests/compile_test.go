package tests

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
)

func TestCompile(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "compile", "compile.svtest"))
}

// go-sieve has more simple error handling, but we still run
// tests to check whether any invalid scripts are not loaded as valid.

func TestCompileErrors(t *testing.T) {
	t.Skip("requires relational extension")
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "compile", "errors.svtest"))
}

func TestCompileRecover(t *testing.T) {
	t.Skip("requires relational extension")
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "compile", "recover.svtest"))
}

func TestCompileWarnings(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "compile", "warnings.svtest"))
}
