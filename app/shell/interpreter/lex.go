package interpreter

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

const (
	eof = -1

	spaceChars = " \t\r\n"
)

type stateFunc func(*lexer) stateFunc

type lexer struct {
	input  string
	state  stateFunc
	tokens chan token

	atEOF bool
	start int
	pos   int
	width int
}

func newLexer(input string) *lexer {
	return &lexer{
		input:  input,
		tokens: make(chan token, 4),
		state:  lexText,
	}
}

func (l *lexer) nextToken() token {
	for {
		select {
		case tok := <-l.tokens:
			return tok
		default:
			if l.atEOF {
				return token{
					typ:     tokenEOF,
					literal: "",
					pos:     len(l.input),
				}
			}
			l.state = l.state(l)
		}
	}
}

func (l *lexer) emit(typ tokenType) {
	if typ == tokenEOF {
		assert.Assert(l.start == l.pos)

		tok := token{typ: typ, pos: len(l.input)}
		l.tokens <- tok
		l.start = l.pos
		return
	}

	assert.Assert(l.pos >= l.start)

	literal := l.input[l.start:l.pos]
	tok := token{
		pos:     l.pos,
		typ:     typ,
		literal: literal,
	}
	l.tokens <- tok
	l.start = l.pos
}

// next advances the lexer one rune and return it
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.atEOF = true
		return eof
	}

	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = width
	l.pos += width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	if !l.atEOF && l.pos > 0 {
		l.pos -= l.width
	}
}

func (l *lexer) discard() {
	l.start = l.pos
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.tokens <- token{
		typ:     tokenError,
		literal: fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) skipWhitespace() {
	for r := l.next(); isSpace(r); {
		r = l.next()
	}
	l.backup()
	l.discard()
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

func lexText(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			_ = l.accept(spaceChars)
			l.emit(tokenSpace)
			l.skipWhitespace()
			return lexText
		case r == '\'':
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return lexSingleQuotes
		case r == '"':
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return lexDoubleQuotes
		case r == eof:
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.emit(tokenEOF)
			return nil
		}
	}
}

func lexSingleQuotes(l *lexer) stateFunc {
	assert.Assert(l.accept("'"))

	l.emit(tokenSingleQuote)

	for {
		switch l.next() {
		case '\'':
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.next()
			l.emit(tokenSingleQuote)
			return lexText
		case eof:
			return l.errorf("unclosed single quotes")
		default:
		}
	}
}

func lexDoubleQuotes(l *lexer) stateFunc {
	assert.Assert(l.accept(`"`))

	l.emit(tokenDoubleQuote)

	for {
		switch l.next() {
		case '"':
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.next()
			l.emit(tokenDoubleQuote)
			return lexText
		case eof:
			return l.errorf("unclosed double quotes")
		default:
		}
	}
}
