package extensions

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve/tests"
)

func TestSubaddressBasic(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "subaddress", "basic.svtest"))
}

func TestSubaddressConfig(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "subaddress", "config.svtest"))
}

func TestSubaddressRFC(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "subaddress", "rfc.svtest"))
}
