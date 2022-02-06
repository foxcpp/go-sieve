package lexer

import (
	"fmt"
)

type Stream struct {
	cursor int
	toks   []Token
}

func (s *Stream) Last() Token {
	if s.cursor >= len(s.toks) {
		return nil
	}
	return s.toks[s.cursor]
}

func (s *Stream) Pop() Token {
	s.cursor++
	if s.cursor >= len(s.toks) {
		return nil
	}
	return s.toks[s.cursor]
}

func (s *Stream) Peek() Token {
	cur := s.cursor + 1
	if cur >= len(s.toks) {
		return nil
	}
	return s.toks[cur]
}

func (s *Stream) Err(format string, args ...interface{}) error {
	last := s.Last()
	if last == nil {
		return fmt.Errorf(format, args...)
	}
	return ErrorAt(last, format, args...)
}

func NewStream(toks []Token) *Stream {
	return &Stream{cursor: -1, toks: toks}
}
