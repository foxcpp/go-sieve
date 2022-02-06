package interp

import (
	"strings"

	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

var supportedExtensions = map[string]struct{}{
	"fileinto": {},
	"envelope": {},
}

var (
	commands map[string]func(*Script, parser.Cmd) (Cmd, error)
	tests    map[string]func(*Script, parser.Test) (Test, error)
)

func init() {
	// break initialization loop

	commands = map[string]func(*Script, parser.Cmd) (Cmd, error){
		// RFC 5228 Actions
		"require": loadRequire,
		"if":      loadIf,
		"elsif":   loadElsif,
		"else":    loadElse,
		"stop":    loadStop,
		// RFC 5228 Actions
		"fileinto": loadFileInto, // fileinto extension
		"redirect": loadRedirect,
		"keep":     loadKeep,
		"discard":  loadDiscard,
	}
	tests = map[string]func(*Script, parser.Test) (Test, error){
		// RFC 5228 Tests
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
	}
}

func LoadScript(cmdStream []parser.Cmd, opts *Options) (*Script, error) {
	s := &Script{
		opts: opts,
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
