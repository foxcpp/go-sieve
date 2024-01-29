package parser

import (
	"github.com/foxcpp/go-sieve/lexer"
)

type Options struct {
	MaxBlockNesting int
	MaxTestNesting  int
}

func Parse(stream *lexer.Stream, opts *Options) ([]Cmd, error) {
	return parse(stream, 0, opts)
}

// parse is a low-level parsing function, it creates
// AST with very little interpretation of values.
func parse(stream *lexer.Stream, nesting int, opts *Options) ([]Cmd, error) {
	if opts.MaxBlockNesting != 0 && nesting > opts.MaxBlockNesting {
		return nil, stream.Err("block nesting limit exceeded")
	}
	res := []Cmd{}
	for {
		curCmd := Cmd{}

		idT := stream.Pop()
		if idT == nil {
			return res, nil
		}
		switch id := idT.(type) {
		case lexer.Identifier:
			curCmd.Id = id.Text
			curCmd.Position = id.Position
		case lexer.BlockEnd:
			return res, nil
		default:
			return nil, stream.Err("reading command: expected an identifier or closing brace")
		}

		args, tests, err := readArguments(stream, false, 0, opts)
		if err != nil {
			return nil, err
		}
		curCmd.Args = args
		curCmd.Tests = tests

		cmdEnd := stream.Pop()
		if cmdEnd == nil {
			return nil, stream.Err("reading command: expected semicolon or block")
		}
		switch cmdEnd.(type) {
		case lexer.Semicolon:
			// Ok.
		case lexer.BlockStart:
			cmds, err := parse(stream, nesting+1, opts)
			if err != nil {
				return nil, err
			}

			// EOF vs } check
			last := stream.Last()
			if last == nil {
				return nil, stream.Err("reading command: expected a closing brace")
			}

			curCmd.Block = cmds
		default:
			return nil, stream.Err("reading command: unexpected token")
		}

		res = append(res, curCmd)
	}
}

func readArguments(s *lexer.Stream, forTest bool, nesting int, opts *Options) ([]Arg, []Test, error) {
	if opts.MaxTestNesting != 0 && nesting > opts.MaxTestNesting {
		return nil, nil, s.Err("reading arguments: nesting limit exceeded")
	}
	var args []Arg
	var tests []Test

	for {
		tok := s.Peek()
		if tok == nil {
			return nil, nil, s.Err("reading arguments: expected semicolon or arguments or block, got EOF")
		}
		switch tok := tok.(type) {
		case lexer.Semicolon, lexer.BlockStart:
			return args, tests, nil
		case lexer.Comma, lexer.TestListEnd:
			if !forTest {
				return nil, nil, s.Err("reading arguments: expected semicolon or arguments or block, got %v", tok)
			}
			return args, tests, nil
		case lexer.String:
			s.Pop()
			args = append(args, StringArg{Value: tok.Text, Position: tok.Position})
		case lexer.ListStart:
			s.Pop()
			list, err := readStringList(s)
			if err != nil {
				return nil, nil, err
			}
			args = append(args, StringListArg{Value: list, Position: tok.Position})
		case lexer.Number:
			s.Pop()
			args = append(args, NumberArg{Value: tok.Value * tok.Quantifier.Multiplier(), Position: tok.Position})
		case lexer.Colon:
			s.Pop() // colon
			idT := s.Pop()
			if idT == nil {
				return nil, nil, s.Err("reading arguments: expected identifier, got EOF")
			}
			id, ok := idT.(lexer.Identifier)
			if !ok {
				return nil, nil, s.Err("reading arguments: expected identifier")
			}
			args = append(args, TagArg{Value: id.Text, Position: tok.Position})
		case lexer.Identifier:
			// a single test, at the end of arguments.
			s.Pop()
			t := Test{
				Position: tok.Position,
				Id:       tok.Text,
			}
			tArgs, tTests, err := readArguments(s, true, nesting+1, opts)
			if err != nil {
				return nil, nil, err
			}

			t.Args = tArgs
			t.Tests = tTests
			tests = append(tests, t)
		case lexer.TestListStart:
			s.Pop()
			var err error
			tests, err = readTestList(s, nesting, opts)
			if err != nil {
				return nil, nil, err
			}
			return args, tests, nil
		default:
			return nil, nil, s.Err("reading arguments: expected semicolon or arguments or block. got %v", tok)
		}
	}
}

func readTestList(s *lexer.Stream, nesting int, opts *Options) ([]Test, error) {
	needTest := true
	res := []Test{}
	for {
		tok := s.Pop()
		if tok == nil {
			return nil, s.Err("reading test list: expected identifier, got EOF")
		}
		switch tok := tok.(type) {
		case lexer.Identifier:
			if !needTest {
				return nil, s.Err("reading test list: expected comma or closing brace, got identifier")
			}
			t := Test{
				Position: tok.Position,
				Id:       tok.Text,
			}
			args, tests, err := readArguments(s, true, nesting+1, opts)
			if err != nil {
				return nil, err
			}
			t.Args = args
			t.Tests = tests
			res = append(res, t)
			needTest = false
		case lexer.Comma:
			if needTest {
				return nil, s.Err("reading test list: expected identifier or list end, got comma")
			}
			needTest = true
		case lexer.TestListEnd:
			return res, nil
		default:
			return nil, s.Err("reading test list: expected identifier or comma or closing brace, got %v", tok)
		}
	}
}

func readStringList(s *lexer.Stream) ([]string, error) {
	res := []string{}
	needString := true
	for {
		tok := s.Pop()
		if tok == nil {
			return nil, s.Err("reading string list: expected string or closing brace, got EOF")
		}
		switch tok := tok.(type) {
		case lexer.String:
			if !needString {
				return nil, s.Err("reading string list: expected comma or closing brace, got string")
			}
			res = append(res, tok.Text)
			needString = false
		case lexer.Comma:
			if needString {
				return nil, s.Err("reading string list: expected string, got comma")
			}
			needString = true
		case lexer.ListEnd:
			return res, nil
		default:
			return nil, s.Err("reading string list: expected string, comma or closing brace")
		}
	}
}

func ErrorAt(pos lexer.Position, fmt string, args ...interface{}) error {
	return lexer.ErrorAt(pos, fmt, args...)
}
