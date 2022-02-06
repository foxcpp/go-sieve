package parser

type Arg interface {
	arg()
}

type NumberArg int

func (NumberArg) arg() {}

type StringArg string

func (StringArg) arg() {}

type StringListArg []string

func (StringListArg) arg() {}

type TagArg string

func (TagArg) arg() {}

type Test struct {
	Id    string
	Args  []Arg
	Tests []Test
}

type Cmd struct {
	Identifier string
	Args       []Arg
	Tests      []Test
	Block      []Cmd
}
