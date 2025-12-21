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

type stateFunc func(*parser) stateFunc

type parser struct {
	state stateFunc
	input string
	err   error

	args []string

	start int
	pos   int
	width int
}

func parseInput(input string) ([]string, error) {
	assert.Assert(len(input) < 1_000_000)

	if len(input) == 0 {
		return nil, nil
	}

	p := &parser{
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
func (p *parser) next() rune {
	if p.pos >= len(p.input) {
		return eof
	}

	r, size := utf8.DecodeRuneInString(p.input[p.pos:])
	p.width = size
	p.pos += size
	return r
}

// backup moves the parser back the width of the latest rune
func (p *parser) backup() {
	p.pos -= p.width
	assert.Assert(p.pos >= 0)
}

// accept advances the parser one rune if the next rune is contained
// in the valid string passed. Checked by strings.ContainsRune
func (p *parser) accept(valid string) bool {
	if p.pos == len(p.input) {
		return false
	}

	if strings.ContainsRune(valid, p.next()) {
		return true
	}
	p.backup()
	return false
}

// acceptRun behaves like accept but continues advancing the parser
// untill no matches are found
func (p *parser) acceptRun(valid string) {
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
func (p *parser) runUntil(valid string) {
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
func (p *parser) discard() {
	p.start = p.pos
}

// skipWhitespace skips past any whitespace defined by ' ', '\t', '\n', and '\r'
func (p *parser) skipWhitespace() {
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
func (p *parser) peak() rune {
	if p.pos == len(p.input) {
		return eof
	}

	r := p.next()
	p.backup()
	return r
}

// addArg adds an arguments to the end of the args list. Noops on empty args
func (p *parser) addArg(arg string) {
	assert.NotNil(p.args)

	if len(arg) == 0 {
		return
	}

	p.args = append(p.args, arg)
}

// current returns the substring of input that is between the start and position of the parser.
// Returns empty string if start and position are the same
func (p *parser) current() string {
	if p.start == p.pos {
		return ""
	}
	cur := p.input[p.start:p.pos]
	p.start = p.pos
	return cur
}

func parseText(p *parser) stateFunc {
	assert.NotNil(p)

	p.skipWhitespace()
	current := ""
	for {
		p.runUntil(whitespace + quotes)
		current += p.current()

		if !p.accept(quotes) {
			break
		}

		if p.accept(quotes) {
			p.discard()
			continue
		} else {
			p.backup()
			p.addArg(current)
			return parseQuoted
		}
	}

	p.addArg(current)

	if p.peak() == eof {
		return nil
	}

	return parseText
}

func parseQuoted(p *parser) stateFunc {
	assert.NotNil(p)

	assert.Assert(p.accept(quotes))
	p.discard()

	current := ""
	for {
		p.runUntil(quotes)
		current += p.current()

		assert.Assert(p.accept(quotes))
		p.discard()

		if !p.accept(quotes) {
			break
		}
		p.discard()
	}

	p.addArg(current)

	if p.peak() == eof {
		return nil
	}

	return parseText
}
