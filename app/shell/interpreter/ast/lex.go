package ast

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

const (
	eof = -1

	spaceChars        = " \t\r\n"
	quotedEscapeChars = `"\$`
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

func (l *lexer) current() rune {
	var r rune
	if l.pos > 0 {
		r, _ = utf8.DecodeRuneInString(l.input[l.pos-l.width:])
	} else {
		r, _ = utf8.DecodeRuneInString(l.input[:])
	}

	return r
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

// func (l *lexer) discard() {
// 	l.start = l.pos
// }

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

// func (l *lexer) skipWhitespace() {
// 	for r := l.next(); isSpace(r); {
// 		r = l.next()
// 	}
// 	l.backup()
// 	l.discard()
// }

func (l *lexer) emitText() {
	if l.pos > l.start {
		l.emit(tokenText)
	}
}

func (l *lexer) escaped() {
	assert.Assert(l.start == l.pos)

	if !l.accept("\\") {
		l.emit(tokenError)
		return
	}

	l.next()
	l.emit(tokenEscaped)
}

func (l *lexer) variable() {
	if !l.accept("$") {
		l.emit(tokenError)
		return
	}

	hasParen := l.accept("{")

	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// continue
		case r == '}':
			if !hasParen {
				l.errorf("unexpected closing paren")
				return
			}

			l.emit(tokenVariable)
			return
		default:
			l.backup()

			if hasParen {
				l.errorf("unclosed variable paren")
				return
			}

			l.emit(tokenVariable)
			return
		}
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func lexText(l *lexer) stateFunc {
	for {
		switch r := l.peek(); {
		case isSpace(r):
			l.emitText()
			l.acceptRun(spaceChars)
			l.emit(tokenSpace)
			return lexText
		case r == '\'':
			l.emitText()
			return lexSingleQuotes
		case r == '"':
			l.emitText()
			l.next()
			l.emit(tokenDoubleQuote)
			return lexInsideDoubleQuotes
		case r == '\\':
			l.emitText()
			l.escaped()
			return lexText
		case r == '|':
			l.emitText()
			l.next()
			l.emit(tokenPipeline)
			return lexText
		case r == eof:
			l.next()
			l.emitText()
			l.emit(tokenEOF)
			return nil
		case r == '&':
			l.emitText()
			l.next()
			l.emit(tokenAmpersand)
			return lexText
		case r == ';':
			l.emitText()
			l.next()
			l.emit(tokenSemicolon)
			return lexText
		case r == '$':
			l.emitText()
			l.variable()
			return lexText
		case r == '>':
			if isSpace(l.current()) {
				return lexRedirectOrAppend
			}

			l.backup()
			if isSpace(l.current()) {
				l.emitText()
				return lexRedirectOrAppend
			}

			return l.errorf("expected space before a redirect at char %d but got \"%c\"", l.pos, l.current())
		default:
			l.next()
		}
	}
}

func lexSingleQuotes(l *lexer) stateFunc {
	assert.Assert(l.accept("'"))

	l.emit(tokenSingleQuote)

	for {
		switch l.peek() {
		case '\'':
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.next()
			l.emit(tokenSingleQuote)
			return lexText
		case eof:
			return l.errorf("unclosed single quotes")
		default:
			l.next()
		}
	}
}

func lexRedirectOrAppend(l *lexer) stateFunc {
	_ = l.accept("12")
	assert.Assert(l.accept(">"))
	if l.accept(">") {
		l.emit(tokenAppend)
	} else {
		l.emit(tokenRedirect)
	}
	return lexText
}

func lexInsideDoubleQuotes(l *lexer) stateFunc {
	for {
		switch l.peek() {
		case '"':
			l.emitText()
			l.next()
			l.emit(tokenDoubleQuote)
			return lexText
		case '$':
			l.emitText()
			l.variable()
			return lexInsideDoubleQuotes
		case '\\':
			l.next()
			if strings.ContainsRune(quotedEscapeChars, l.peek()) {
				l.backup()
				l.emitText()

				l.accept(`\`)
				l.next()
				l.emit(tokenEscaped)
			}

			return lexInsideDoubleQuotes
		case eof:
			return l.errorf("unclosed double quotes")
		default:
			l.next()
		}
	}
}
