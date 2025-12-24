package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextToken(t *testing.T) {
	tt := []struct {
		input  string
		output []token
	}{
		{
			input: " one two ",
			output: []token{
				{tokenSpace, " ", -1},
				{tokenText, "one", -1},
				{tokenSpace, " ", -1},
				{tokenText, "two", -1},
				{tokenSpace, " ", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: "'one   two'",
			output: []token{
				{tokenSingleQuote, "'", -1},
				{tokenText, "one   two", -1},
				{tokenSingleQuote, "'", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: "one   two",
			output: []token{
				{tokenText, "one", -1},
				{tokenSpace, " ", -1},
				{tokenText, "two", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: "'one''two'",
			output: []token{
				{tokenSingleQuote, "'", -1},
				{tokenText, "one", -1},
				{tokenSingleQuote, "'", -1},
				{tokenSingleQuote, "'", -1},
				{tokenText, "two", -1},
				{tokenSingleQuote, "'", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: "one''two",
			output: []token{
				{tokenText, "one", -1},
				{tokenSingleQuote, "'", -1},
				{tokenSingleQuote, "'", -1},
				{tokenText, "two", -1},
				{tokenEOF, "", -1},
			},
		},
	}

	for _, test := range tt {
		t.Run(test.input, func(t *testing.T) {
			lexer := newLexer(test.input)
			for i, outTok := range test.output {
				tok := lexer.nextToken()
				assert.Equal(t, outTok.typ, tok.typ, "token idx %d", i)
				assert.Equal(t, outTok.literal, tok.literal, "token idx %d", i)
			}
		})
	}
}
