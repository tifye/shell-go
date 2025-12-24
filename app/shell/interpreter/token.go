package interpreter

import (
	"fmt"
)

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenSpace
	tokenText
	tokenSingleQuote
	tokenDoubleQuote
)

type token struct {
	typ     tokenType
	literal string
	pos     int
}

func (t token) String() string {
	switch t.typ {
	case tokenEOF:
		return "eof"
	case tokenError:
		return "error"
	}
	if len(t.literal) > 10 {
		return fmt.Sprintf("%.10q...", t.literal)
	}
	return fmt.Sprintf("%q", t.literal)
}
