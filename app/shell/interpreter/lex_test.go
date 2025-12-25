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
		{
			input: `"one   two"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "one   two", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `"one""two"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "one", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "two", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `"one" "two"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "one", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenSpace, ` `, -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "two", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `"one's two"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, `one's two`, -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `three\ \ \ spaces`,
			output: []token{
				{tokenText, "three", -1},
				{tokenEscaped, " ", -1},
				{tokenEscaped, " ", -1},
				{tokenEscaped, " ", -1},
				{tokenText, "spaces", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `one\   two`,
			output: []token{
				{tokenText, "one", -1},
				{tokenEscaped, " ", -1},
				{tokenSpace, " ", -1},
				{tokenText, "two", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `one\ntwo`,
			output: []token{
				{tokenText, "one", -1},
				{tokenEscaped, "n", -1},
				{tokenText, "two", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `one\\two`,
			output: []token{
				{tokenText, "one", -1},
				{tokenEscaped, `\`, -1},
				{tokenText, "two", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `\'one\'`,
			output: []token{
				{tokenEscaped, "'", -1},
				{tokenText, "one", -1},
				{tokenEscaped, "'", -1},
				{tokenEOF, "", -1},
			},
		},
		{
			input: `"A \\ escapes itself"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, "A ", -1},
				{tokenEscaped, `\`, -1},
				{tokenText, " escapes itself", -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, ``, -1},
			},
		},
		{
			input: `"A \" inside double quotes"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, `A `, -1},
				{tokenEscaped, `"`, -1},
				{tokenText, ` inside double quotes`, -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, ``, -1},
			},
		},
		{
			input: `"\9"`,
			output: []token{
				{tokenDoubleQuote, `"`, -1},
				{tokenText, `\9`, -1},
				{tokenDoubleQuote, `"`, -1},
				{tokenEOF, ``, -1},
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
