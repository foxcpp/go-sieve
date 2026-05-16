package extensions

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve/tests"
)

func TestIMAP4FlagsBasic(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "imap4flags", "basic.svtest"))
}

func TestIMAP4FlagsErrors(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "imap4flags", "errors.svtest"))
}

func TestIMAP4FlagsFlagStore(t *testing.T) {
	t.Skip("Unsupported extension: mailbox")
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "imap4flags", "flagstore.svtest"))
}

func TestIMAP4FlagsFlagString(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "imap4flags", "flagstring.svtest"))
}

func TestIMAP4FlagsHasFlag(t *testing.T) {
	tests.RunDovecotTest(t, filepath.Join("..", "pigeonhole", "tests", "extensions", "imap4flags", "hasflag.svtest"))
}
