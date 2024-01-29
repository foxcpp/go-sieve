package tests

import (
	"path/filepath"
	"testing"
)

func TestExtensionsEncodedCharacters(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "encoded-character.svtest"))
}
