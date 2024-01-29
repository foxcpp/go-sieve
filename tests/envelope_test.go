package tests

import (
	"path/filepath"
	"testing"
)

func TestExtensionsEnvelope(t *testing.T) {
	RunDovecotTestWithout(t, filepath.Join("pigeonhole", "tests", "extensions", "envelope.svtest"),
		[]string{
			// Parser does not understand source routes
			"Envelope - source route",
			"Envelope - source route errors",
			// Envelope address validation is left to the library user e.g. SMTP server.
			"Envelope - invalid paths",
			"Envelope - syntax errors",
		})
}
