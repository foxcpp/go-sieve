package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Write(w io.Writer, toks []Token) error {
	bw := bufio.NewWriter(w)
	for _, t := range toks {
		var err error
		switch t := t.(type) {
		case Identifier:
			_, err = bw.WriteString(t.Text)
		case Number:
			if t.Quantifier != None {
				_, err = fmt.Fprintf(bw, "%d%s", t.Value, string(t.Quantifier))
			} else {
				_, err = fmt.Fprintf(bw, "%d", t.Value)
			}
		case String:
			_, err = bw.WriteString(formatString(t.Text))
		case ListStart:
			err = bw.WriteByte('[')
		case ListEnd:
			err = bw.WriteByte(']')
		case TestListStart:
			err = bw.WriteByte('(')
		case TestListEnd:
			err = bw.WriteByte(')')
		case BlockStart:
			err = bw.WriteByte('{')
		case BlockEnd:
			err = bw.WriteByte('}')
		case Comma:
			err = bw.WriteByte(',')
		case Semicolon:
			err = bw.WriteByte(';')
		case Colon:
			err = bw.WriteByte(':')
		default:
			panic("unexpected token type")
		}
		if err != nil {
			return err
		}

		// TODO: Preserve whitespace properly instead?
		if err := bw.WriteByte(' '); err != nil {
			return err
		}
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}

func formatString(s string) string {
	esc := strings.Builder{}
	esc.WriteByte('"')
	for _, r := range []byte(s) {
		switch r {
		case '"':
			esc.WriteString(`\"`)
		case '\\':
			esc.WriteString(`\\`)
		default:
			esc.WriteByte(r)
		}
	}
	esc.WriteByte('"')
	return esc.String()
}
