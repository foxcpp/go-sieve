package tests

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
)

func TestExtensionsEnvelope(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "envelope.svtest"))
}
