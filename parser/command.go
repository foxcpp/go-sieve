package parser

import "github.com/foxcpp/go-sieve/lexer"

type Arg interface {
	LineCol() (int, int)
	arg()
}

type NumberArg struct {
	Value int
	lexer.Position
}

func (NumberArg) arg() {}

type StringArg struct {
	Value string
	lexer.Position
}

func (StringArg) arg() {}

type StringListArg struct {
	Value []string
	lexer.Position
}

func (StringListArg) arg() {}

type TagArg struct {
	Value string
	lexer.Position
}

func (TagArg) arg() {}

type Test struct {
	lexer.Position
	Id    string
	Args  []Arg
	Tests []Test
}

type Cmd struct {
	lexer.Position
	Id    string
	Args  []Arg
	Tests []Test
	Block []Cmd
}
