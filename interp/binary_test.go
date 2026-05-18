package interp

import "testing"

func TestSaveRestorePreservesSubAddressSep(t *testing.T) {
	script := Script{
		extensions: map[string]struct{}{"subaddress": {}},
		opts: &Options{
			MaxRedirects:       5,
			MaxVariableCount:   128,
			MaxVariableNameLen: 32,
			MaxVariableLen:     4000,
			SubAddressSep:      "-",
		},
		cmd: []Cmd{},
	}

	blob, err := script.Save()
	if err != nil {
		t.Fatal(err)
	}

	restored, err := Restore(blob)
	if err != nil {
		t.Fatal(err)
	}

	if restored.opts.SubAddressSep != "-" {
		t.Fatalf("SubAddressSep not preserved, got %q", restored.opts.SubAddressSep)
	}
}
