package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emersion/go-message/textproto"
	"github.com/foxcpp/go-sieve"
	"github.com/foxcpp/go-sieve/interp"
)

func DefaultTestRuntime(d *interp.RuntimeData) {
	d.Test = &interp.TestRuntime{}
}

func ExecuteTestRuntime(env interp.ExecuteTestEnvironment) func(d *interp.RuntimeData) {
	if env == nil {
		panic("environment must be provided")
	}
	return func(d *interp.RuntimeData) {
		if d.Test == nil {
			d.Test = &interp.TestRuntime{}
		}
		d.Test.Execute = env
	}
}

func RunDovecotTestInline(t *testing.T, baseDir string, scriptText string, testParams ...func(d *interp.RuntimeData)) {
	t.Helper()

	opts := sieve.DefaultOptions()
	opts.Lexer.Filename = "inline"
	opts.Interp.T = t

	script, err := sieve.Load(strings.NewReader(scriptText), opts)
	if err != nil {
		t.Fatal("Loading test script:", err)
	}

	ctx := context.Background()

	// Empty data.
	data := sieve.NewRuntimeData(script, interp.DummyPolicy{},
		interp.EnvelopeStatic{}, interp.MessageStatic{
			Size:   0,
			Header: &textproto.Header{},
		})
	data.Test = &interp.TestRuntime{
		Name: t.Name(),
	}
	for _, param := range testParams {
		param(data)
	}

	if baseDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		data.Namespace = os.DirFS(wd)
	} else {
		data.Namespace = os.DirFS(baseDir)
	}

	err = script.Execute(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
}

func RunDovecotTestWithout(t *testing.T, path string, disabledTests []string, testParams ...func(d *interp.RuntimeData)) {
	t.Helper()

	svScript, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	opts := sieve.DefaultOptions()
	opts.Lexer.Filename = filepath.Base(path)
	opts.Interp.T = t
	opts.Interp.DisabledTests = disabledTests

	script, err := sieve.Load(bytes.NewReader(svScript), opts)
	if err != nil {
		t.Fatal("Loading test script:", err)
	}

	ctx := context.Background()

	// Empty data.
	data := sieve.NewRuntimeData(script, interp.DummyPolicy{},
		interp.EnvelopeStatic{}, interp.MessageStatic{
			Size:   0,
			Header: &textproto.Header{},
		})
	data.Namespace = os.DirFS(filepath.Dir(path))
	data.Test = &interp.TestRuntime{
		Name: t.Name(),
	}
	// Provide default environment items for the environment extension (RFC 5183).
	// These values match what Pigeonhole's testsuite expects (basic.svtest, rfc.svtest).
	data.SieveEnv = interp.MapSieveEnvironment{
		"name":     "Pigeonhole Sieve",
		"version":  "0.0",
		"location": "MS",
		"phase":    "during",
	}
	for _, param := range testParams {
		param(data)
	}

	err = script.Execute(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
}

func RunDovecotTest(t *testing.T, path string, opts ...func(d *interp.RuntimeData)) {
	t.Helper()
	RunDovecotTestWithout(t, path, nil, opts...)
}
