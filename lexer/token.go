package lexer

import (
	"fmt"
	"strconv"
)

type Position struct {
	File string
	Line int
	Col  int
}

func (l Position) String() string {
	if l.File != "" {
		return l.File + ":" + strconv.Itoa(l.Line) + ":" + strconv.Itoa(l.Col)
	}
	return strconv.Itoa(l.Line) + ":" + strconv.Itoa(l.Col)
}

func (l Position) LineCol() (int, int) {
	return l.Line, l.Col
}

func LineCol(line, col int) Position {
	return Position{Line: line, Col: col}
}

type Token interface {
	LineCol() (int, int)
	String() string
}

type Identifier struct {
	Position
	Text string
}

func (t Identifier) String() string { return fmt.Sprintf(`Identifiner("%s")`, t.Text) }

type Quantifier byte

const (
	None Quantifier = '\x00'
	Kilo Quantifier = 'K'
	Mega Quantifier = 'M'
	Giga Quantifier = 'G'
)

func (q Quantifier) Multiplier() int {
	switch q {
	case None:
		return 1
	case Kilo:
		return 1024
	case Mega:
		return 1024 * 1024
	case Giga:
		return 1024 * 1024 * 1024
	default:
		panic("unknown quantifier")
	}
}

type Number struct {
	Position
	Value      int
	Quantifier Quantifier
}

func (t Number) String() string {
	if t.Quantifier != None {
		return fmt.Sprintf("Number(%d, %v)", t.Value, string(t.Quantifier))
	}
	return fmt.Sprintf("Number(%d)", t.Value)

}

type String struct {
	Position
	Text string
}

func (t String) String() string { return fmt.Sprintf(`String("%s")`, t.Text) }

type ListStart struct{ Position }

func (ListStart) String() string { return "ListStart()" }

type ListEnd struct{ Position }

func (ListEnd) String() string { return "ListEnd()" }

type BlockStart struct{ Position }

func (BlockStart) String() string { return "BlockStart()" }

type BlockEnd struct{ Position }

func (BlockEnd) String() string { return "BlockEnd()" }

type TestListStart struct{ Position }

func (TestListStart) String() string { return "TestListStart()" }

type TestListEnd struct{ Position }

func (TestListEnd) String() string { return "TestListEnd()" }

type Comma struct{ Position }

func (Comma) String() string { return "Comma()" }

type Semicolon struct{ Position }

func (Semicolon) String() string { return "Semicolon()" }

type Colon struct{ Position }

func (Colon) String() string { return "Colon()" }

type position interface {
	LineCol() (int, int)
}

type tokError struct {
	t    position
	text string
}

func (e tokError) Error() string {
	if e.t == nil {
		return fmt.Sprintf("unknown-position: %s", e.text)
	}
	line, col := e.t.LineCol()
	if line == 0 || col == 0 {
		return fmt.Sprintf("invalid-position: %s", e.text)
	}
	return fmt.Sprintf("%d:%d: %s", line, col, e.text)
}

func ErrorAt(t position, format string, args ...interface{}) error {
	return tokError{t: t, text: fmt.Sprintf(format, args...)}
}
