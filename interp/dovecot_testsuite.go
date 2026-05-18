package interp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
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
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}

	testData := d.Copy()
	testData.Test.Name = c.TestName
	testData.Test.FailMessage = ""

	d.Script.opts.T.Run(c.TestName, func(t *testing.T) {
		for _, testName := range testData.Script.opts.DisabledTests {
			if c.TestName == testName {
				t.Skip("test is disabled by DisabledTests")
			}
		}

		for _, cmd := range c.Cmds {
			if err := cmd.Execute(ctx, testData); err != nil {
				if errors.Is(err, ErrStop) {
					if testData.Test.FailMessage != "" {
						t.Errorf("test_fail at %v called: %v", testData.Test.FailAt, testData.Test.FailMessage)
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
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}

	d.Test.FailMessage = expandVars(d, c.Message)
	d.Test.FailAt = c.At
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

type CmdDovecotBinarySave struct {
	Name string
}

func (c CmdDovecotBinarySave) Execute(_ context.Context, d *RuntimeData) error {
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}
	if d.Test.SavedScripts == nil {
		d.Test.SavedScripts = make(map[string][]byte)
	}
	if d.Test.Script == nil {
		return fmt.Errorf("no script loaded to save")
	}

	blob, err := d.Test.Script.Save()
	if err != nil {
		return fmt.Errorf("failed to encode script: %v", err)
	}

	d.Test.SavedScripts[c.Name] = blob
	return nil
}

type CmdDovecotMessage struct {
	SMTP   bool
	Folder string
	Index  int
}

func (c CmdDovecotMessage) Execute(_ context.Context, d *RuntimeData) error {
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Execute == nil {
		return fmt.Errorf("no test execution environment is configured")
	}

	// Implicit test_dovecot_result_reset
	if d.Test.OriginalMsg != nil {
		d.Msg = d.Test.OriginalMsg
		d.Envelope = d.Test.OriginalEnvelope
	}

	d.Test.OriginalMsg = d.Msg
	d.Test.OriginalEnvelope = d.Envelope
	d.Test.OriginalFlags = d.Flags

	if c.SMTP {
		msg, err := d.Test.Execute.GetSMTPMessage(c.Index)
		if err != nil {
			return err
		}
		d.Msg = msg.Message
		d.Envelope = msg.Envelope
		return nil
	}

	folder := c.Folder
	if folder == "" {
		folder = d.Test.Execute.GetDefaultMailbox()
	}

	msg, err := d.Test.Execute.GetMailboxMessage(folder, c.Index)
	if err != nil {
		return err
	}
	d.Msg = msg.Message
	d.Envelope = msg.Envelope
	d.Flags = msg.Flags

	return nil
}

type CmdDovecotMailboxCreate struct {
	Name string
}

func (c CmdDovecotMailboxCreate) Execute(_ context.Context, d *RuntimeData) error {
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Execute == nil {
		return fmt.Errorf("no test execution environment is configured")
	}

	return d.Test.Execute.CreateMailbox(c.Name)
}

type TestDovecotMessage struct {
	SMTP   bool
	Folder string
	Index  int
}

func (c TestDovecotMessage) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if d.Test == nil {
		return false, fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Execute == nil {
		return false, fmt.Errorf("no test execution environment is configured")
	}

	if c.SMTP {
		return d.Test.Execute.HasSMTPMessage(c.Index)
	}

	folder := c.Folder
	if folder == "" {
		folder = d.Test.Execute.GetDefaultMailbox()
	}

	return d.Test.Execute.HasMailboxMessage(folder, c.Index)
}

type CmdDovecotResultReset struct{}

func (c CmdDovecotResultReset) Execute(_ context.Context, d *RuntimeData) error {
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Execute == nil {
		return fmt.Errorf("no test execution environment is configured")
	}

	if d.Test.OriginalMsg == nil {
		return nil
	}

	d.Envelope = d.Test.OriginalEnvelope
	d.Msg = d.Test.OriginalMsg
	d.Flags = d.Test.OriginalFlags

	return nil
}

type CmdDovecotBinaryLoad struct {
	Name string
}

func (c CmdDovecotBinaryLoad) Execute(_ context.Context, d *RuntimeData) error {
	if d.Test == nil {
		return fmt.Errorf("test runtime is not configured")
	}

	blob := d.Test.SavedScripts[c.Name]
	if blob == nil {
		return fmt.Errorf("no such script loaded")
	}

	restored, err := Restore(blob)
	if err != nil {
		return err
	}

	d.Test.Script = restored
	return nil
}

type TestDovecotCompile struct {
	ScriptPath string
}

func (t TestDovecotCompile) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if d.Test == nil {
		return false, fmt.Errorf("test runtime is not configured")
	}
	if d.Namespace == nil {
		return false, fmt.Errorf("RuntimeData.Namespace is not set, cannot load scripts")
	}

	svScript, err := fs.ReadFile(d.Namespace, t.ScriptPath)
	if err != nil {
		d.Script.opts.T.Log("ReadFile failed:", err)
		return false, nil
	}

	toks, err := lexer.Lex(bytes.NewReader(svScript), &lexer.Options{
		Filename:  t.ScriptPath,
		MaxTokens: 5000,
	})
	if err != nil {
		d.Script.opts.T.Log("lexer.Lex failed:", err)
		return false, nil
	}

	cmds, err := parser.Parse(lexer.NewStream(toks), &parser.Options{
		MaxBlockNesting: d.Test.MaxNesting,
		MaxTestNesting:  d.Test.MaxNesting,
	})
	if err != nil {
		d.Script.opts.T.Log("parser.Parse failed:", err)
		return false, nil
	}

	script, err := LoadScript(cmds, &Options{
		MaxRedirects: d.Script.opts.MaxRedirects,
	})
	if err != nil {
		d.Script.opts.T.Log("LoadScript failed:", err)
		return false, nil
	}

	d.Test.Script = script
	return true, nil
}

type TestDovecotRun struct {
}

func (t TestDovecotRun) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	if d.Test == nil {
		return false, fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Script == nil {
		return false, nil
	}

	testD := d.Copy()
	testD.Script = d.Test.Script
	// Note: Loaded script has no test environment available -
	// it is a regular Sieve script.

	err := d.Test.Script.Execute(ctx, testD)
	if err != nil {
		return false, nil
	}

	// Copy actions into test case RuntimeData so test_results_execute
	// can see it and act on it.
	d.AppliedActions = testD.AppliedActions

	return true, nil
}

type TestDovecotTestError struct {
	matcherTest
}

func (t TestDovecotTestError) Check(_ context.Context, _ *RuntimeData) (bool, error) {
	// go-sieve has a very different error formatting and stops lexing/parsing/loading
	// on first error, therefore we skip all test_errors checks as they are
	// Pigeonhole-specific.
	return true, nil
}

type TestDovecotResultAction struct {
	matcherTest
	Index *int
}

func (t TestDovecotResultAction) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if t.isCount() {
		entryCount := uint64(0)
		if t.Index != nil {
			// Pigeonhole uses 1-based indexing for :index in test_result_action.
			idx := *t.Index - 1
			if idx >= 0 && idx < len(d.AppliedActions) {
				entryCount++
			}
		} else {
			entryCount = uint64(len(d.AppliedActions))
		}

		return t.countMatches(d, entryCount), nil
	}

	if t.Index != nil {
		// Pigeonhole uses 1-based indexing for :index in test_result_action.
		idx := *t.Index - 1
		if idx < 0 || idx >= len(d.AppliedActions) {
			return false, nil
		}
		action := d.AppliedActions[idx]

		ok, err := t.matcherTest.tryMatch(d, action.testActionName())
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	for _, action := range d.AppliedActions {
		ok, err := t.matcherTest.tryMatch(d, action.testActionName())
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type TestDovecotResultExecute struct{}

func (t TestDovecotResultExecute) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if d.Test == nil {
		return false, fmt.Errorf("test runtime is not configured")
	}
	if d.Test.Execute == nil {
		return false, fmt.Errorf("test execution environment is not configured")
	}

	if err := d.Test.Execute.ExecuteActions(d, d.AppliedActions); err != nil {
		return false, err
	}
	return true, nil
}

func init() {
	gob.Register(CmdDovecotTest{})
	gob.Register(CmdDovecotTestFail{})
	gob.Register(CmdDovecotConfigSet{})
	gob.Register(CmdDovecotBinarySave{})
	gob.Register(CmdDovecotBinaryLoad{})
	gob.Register(CmdDovecotMessage{})
	gob.Register(CmdDovecotResultReset{})
	gob.Register(CmdDovecotMailboxCreate{})
	gob.Register(TestDovecotMessage{})
	gob.Register(TestDovecotCompile{})
	gob.Register(TestDovecotRun{})
	gob.Register(TestDovecotTestError{})
	gob.Register(TestDovecotResultAction{})
	gob.Register(TestDovecotResultExecute{})
}
