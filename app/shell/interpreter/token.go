package interpreter

import (
	"fmt"
)

//go:generate stringer -type tokenType -trimprefix token
type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenSpace
	tokenText
	tokenSingleQuote
	tokenDoubleQuote
	tokenEscaped
	tokenRedirect
	tokenAppend
	tokenPipeline
	tokenAmpersand
	tokenVariable
	tokenSemicolon
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
