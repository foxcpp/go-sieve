package interp

import (
	"context"
	"strings"

	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

var supportedRequires = map[string]struct{}{
	"fileinto":          {},
	"envelope":          {},
	"encoded-character": {},

	"comparator-i;octet":           {},
	"comparator-i;ascii-casemap":   {},
	"comparator-i;ascii-numeric":   {},
	"comparator-i;unicode-casemap": {},

	"imap4flags": {},
	"variables":  {},
	"relational": {},
}

var (
	commands map[string]func(*Script, parser.Cmd) (Cmd, error)
	tests    map[string]func(*Script, parser.Test) (Test, error)
)

func init() {
	// break initialization loop

	commands = map[string]func(*Script, parser.Cmd) (Cmd, error){
		// RFC 5228
		"require":  loadRequire,
		"if":       loadIf,
		"elsif":    loadElsif,
		"else":     loadElse,
		"stop":     loadStop,
		"fileinto": loadFileInto, // fileinto extension
		"redirect": loadRedirect,
		"keep":     loadKeep,
		"discard":  loadDiscard,
		// RFC 5232 (imap4flags extension)
		"setflag":    loadSetFlag,
		"addflag":    loadAddFlag,
		"removeflag": loadRemoveFlag,
		// RFC 5229 (variables extension)
		"set": loadSet,
		// vnd.dovecot.testsuite
		"test":             loadDovecotTest,
		"test_set":         loadDovecotTestSet,
		"test_fail":        loadDovecotTestFail,
		"test_binary_load": loadNoop, // go-sieve has no intermediate binary representation
		"test_binary_save": loadNoop, // go-sieve has no intermediate binary representation
		// "test_result_execute" // apply script results (validated using test_message)
		// "test_mailbox_create"
		// "test_imap_metadata_set"
		"test_config_reload": loadNoop, // go-sieve applies changes immediately
		"test_config_set":    loadDovecotConfigSet,
		"test_config_unset":  loadDovecotConfigUnset,
		// "test_result_reset"
		// "test_message"

	}
	tests = map[string]func(*Script, parser.Test) (Test, error){
		// RFC 5228
		"address":  loadAddressTest,
		"allof":    loadAllOfTest,
		"anyof":    loadAnyOfTest,
		"envelope": loadEnvelopeTest, // envelope extension
		"exists":   loadExistsTest,
		"false":    loadFalseTest,
		"true":     loadTrueTest,
		"header":   loadHeaderTest,
		"not":      loadNotTest,
		"size":     loadSizeTest,
		// RFC 5229 (variables extension)
		"string": loadStringTest,
		// vnd.dovecot.testsuite
		"test_script_compile": loadDovecotCompile, // compile script (to test for compile errors)
		"test_script_run":     loadDovecotRun,     // run script (to test for run-time errors)
		"test_error":          loadDovecotError,   // check detailed results of test_script_compile or test_script_run
		// "test_message" // check results of test_result_execute - where messages are
		// "test_result_action" // check results of test_result_execute - what actions are executed
		// "test_result_reset" // clean results as observed by test_result_action
	}
}

func LoadScript(cmdStream []parser.Cmd, opts *Options) (*Script, error) {
	s := &Script{
		extensions: map[string]struct{}{},
		opts:       opts,
	}

	loadedCmds, err := LoadBlock(s, cmdStream)
	if err != nil {
		return nil, err
	}
	s.cmd = loadedCmds

	return s, nil
}

func LoadBlock(s *Script, cmds []parser.Cmd) ([]Cmd, error) {
	loaded := make([]Cmd, 0, len(cmds))
	for _, c := range cmds {
		cmd, err := LoadCmd(s, c)
		if err != nil {
			return nil, err
		}
		if cmd == nil {
			continue
		}
		loaded = append(loaded, cmd)
	}
	return loaded, nil
}

func LoadCmd(s *Script, cmd parser.Cmd) (Cmd, error) {
	cmdName := strings.ToLower(cmd.Id)
	factory := commands[cmdName]
	if factory == nil {
		return nil, lexer.ErrorAt(cmd, "LoadBlock: unsupported command: %v", cmdName)
	}
	return factory(s, cmd)

}

func LoadTest(s *Script, t parser.Test) (Test, error) {
	testName := strings.ToLower(t.Id)
	factory := tests[testName]
	if factory == nil {
		return nil, lexer.ErrorAt(t, "LoadTest: unsupported test: %v", testName)
	}
	return factory(s, t)
}

type CmdNoop struct{}

func (c CmdNoop) Execute(_ context.Context, _ *RuntimeData) error {
	return nil
}

func loadNoop(_ *Script, _ parser.Cmd) (Cmd, error) {
	return CmdNoop{}, nil
}
