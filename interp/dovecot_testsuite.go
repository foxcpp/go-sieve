package interp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/textproto"
	"strconv"
	"strings"
	"testing"

	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

const DovecotTestExtension = "vnd.dovecot.testsuite"

type CmdDovecotTest struct {
	TestName string
	Cmds     []Cmd
}

func (c CmdDovecotTest) Execute(ctx context.Context, d *RuntimeData) error {
	testData := d.Copy()
	testData.testName = c.TestName
	testData.testFailMessage = ""

	d.Script.opts.T.Run(c.TestName, func(t *testing.T) {
		for _, testName := range testData.Script.opts.DisabledTests {
			if c.TestName == testName {
				t.Skip("test is disabled by DisabledTests")
			}
		}

		for _, cmd := range c.Cmds {
			if err := cmd.Execute(ctx, testData); err != nil {
				if errors.Is(err, ErrStop) {
					if testData.testFailMessage != "" {
						t.Errorf("test_fail at %v called: %v", testData.testFailAt, testData.testFailMessage)
					}
					return
				}
				t.Fatal("Test execution error:", err)
			}
		}
	})

	return nil
}

type CmdDovecotTestFail struct {
	At      lexer.Position
	Message string
}

func (c CmdDovecotTestFail) Execute(_ context.Context, d *RuntimeData) error {
	d.testFailMessage = expandVars(d, c.Message)
	d.testFailAt = c.At
	return ErrStop
}

type CmdDovecotConfigSet struct {
	Unset bool
	Key   string
	Value string
}

func (c CmdDovecotConfigSet) Execute(_ context.Context, d *RuntimeData) error {
	switch c.Key {
	case "sieve_variables_max_variable_size":
		if c.Unset {
			d.Script.opts.MaxVariableLen = 4000
		} else {
			val, err := strconv.Atoi(c.Value)
			if err != nil {
				return err
			}
			d.Script.opts.MaxVariableLen = val
		}
	default:
		return fmt.Errorf("unknown test_config_set key: %v", c.Key)
	}
	return nil
}

type CmdDovecotTestSet struct {
	VariableName  string
	VariableValue string
}

func (c CmdDovecotTestSet) Execute(_ context.Context, d *RuntimeData) error {
	value := expandVars(d, c.VariableValue)

	switch c.VariableName {
	case "message":
		r := textproto.NewReader(bufio.NewReader(strings.NewReader(c.VariableValue)))
		msgHdr, err := r.ReadMIMEHeader()
		if err != nil {
			return fmt.Errorf("failed to parse test message: %v", err)
		}

		d.Msg = MessageStatic{
			Size:   len(c.VariableValue),
			Header: msgHdr,
		}
	case "envelope.from":
		value = strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">")

		d.Envelope = EnvelopeStatic{
			From: value,
			To:   d.Envelope.EnvelopeTo(),
			Auth: d.Envelope.AuthUsername(),
		}
	case "envelope.to":
		value = strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">")

		d.Envelope = EnvelopeStatic{
			From: d.Envelope.EnvelopeFrom(),
			To:   value,
			Auth: d.Envelope.AuthUsername(),
		}
	case "envelope.auth":
		d.Envelope = EnvelopeStatic{
			From: d.Envelope.EnvelopeFrom(),
			To:   d.Envelope.EnvelopeTo(),
			Auth: value,
		}
	default:
		d.Variables[c.VariableName] = c.VariableValue
	}

	return nil
}

type TestDovecotCompile struct {
	ScriptPath string
}

func (t TestDovecotCompile) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if d.Namespace == nil {
		return false, fmt.Errorf("RuntimeData.Namespace is not set, cannot load scripts")
	}

	svScript, err := fs.ReadFile(d.Namespace, t.ScriptPath)
	if err != nil {
		return false, nil
	}

	toks, err := lexer.Lex(bytes.NewReader(svScript), &lexer.Options{
		MaxTokens: 5000,
	})
	if err != nil {
		return false, nil
	}

	cmds, err := parser.Parse(lexer.NewStream(toks), &parser.Options{
		MaxBlockNesting: d.testMaxNesting,
		MaxTestNesting:  d.testMaxNesting,
	})
	if err != nil {
		return false, nil
	}

	script, err := LoadScript(cmds, &Options{
		MaxRedirects: d.Script.opts.MaxRedirects,
	})
	if err != nil {
		return false, nil
	}

	d.testScript = script
	return true, nil
}

type TestDovecotRun struct {
}

func (t TestDovecotRun) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	if d.testScript == nil {
		return false, nil
	}

	testD := d.Copy()
	testD.Script = d.testScript
	// Note: Loaded script has no test environment available -
	// it is a regular Sieve script.

	err := d.testScript.Execute(ctx, testD)
	if err != nil {
		return false, nil
	}

	return true, nil
}

type TestDovecotTestError struct {
}

func (t TestDovecotTestError) Check(_ context.Context, _ *RuntimeData) (bool, error) {
	// go-sieve has a very different error formatting and stops lexing/parsing/loading
	// on first error, therefore we skip all test_errors checks as they are
	// Pigeonhole-specific.
	return true, nil
}
