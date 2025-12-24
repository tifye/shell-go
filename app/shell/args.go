package shell

import (
	"strings"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

const (
	eof        rune   = -1
	whitespace string = " \t\n\r"
	quotes     string = `"'`
)

type stateFunc func(*lexer) stateFunc

type lexer struct {
	state stateFunc
	input string
	err   error

	args []string

	start int
	pos   int
	width int
}

func parseInput(input string) ([]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	p := &lexer{
		state: parseText,
		input: input,
		args:  []string{},
	}
	for p.state != nil {
		p.state = p.state(p)
	}
	return p.args, p.err
}

// next advances the parser one rune
func (p *lexer) next() rune {
	if p.pos >= len(p.input) {
		return eof
	}

	r, size := utf8.DecodeRuneInString(p.input[p.pos:])
	p.width = size
	p.pos += size
	return r
}

// backup moves the parser back the width of the latest rune
func (p *lexer) backup() {
	p.pos -= p.width
	assert.Assert(p.pos >= 0)
}

// accept advances the parser one rune if the next rune is contained
// in the valid string passed. Checked by strings.ContainsRune
func (p *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, p.next()) {
		return true
	}
	p.backup()
	return false
}

// acceptRun behaves like accept but continues advancing the parser
// untill no matches are found
func (p *lexer) acceptRun(valid string) {
	next := p.next()
	for {
		if !strings.ContainsRune(valid, next) {
			break
		}

		if next == eof {
			return
		}

		next = p.next()
	}
	p.backup()
}

// runUntil advances the parser until it encounters a rune contained in valid.
// parser stops just before the matched rune
func (p *lexer) runUntil(valid string) {
	next := p.next()
	for {
		if strings.ContainsRune(valid, next) {
			break
		}

		if next == eof {
			return
		}

		next = p.next()
	}
	p.backup()
}

// discard sets the starting position to the current position
// of the parser effectively discard any input inbetween
func (p *lexer) discard() {
	p.start = p.pos
}

// skipWhitespace skips past any whitespace defined by ' ', '\t', '\n', and '\r'
func (p *lexer) skipWhitespace() {
	r := p.next()
	for r == ' ' || r == '\t' || r == '\n' || r == '\r' {
		r = p.next()
	}

	if r != eof {
		p.backup()
	}

	p.discard()
}

// peak returns the next rune in input without advancing the parser
func (p *lexer) peak() rune {
	r := p.next()
	p.backup()
	return r
}

func parseText(p *lexer) stateFunc {
	return nil
}
