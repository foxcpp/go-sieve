package extensions

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve/tests"
)

func TestExtEnvironmentBasic(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "environment", "basic.svtest"))
}

func TestExtEnvironmentRFC(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "environment", "rfc.svtest"))
}
