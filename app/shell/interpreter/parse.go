package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

var (
	ErrCommandNotFound = errors.New("command not found")
)

func Parse(input string, cmdLookup CommandLookuper) (*Program, error) {
	l := newLexer(input)
	p := newParser(l, cmdLookup)
	prog := p.parse()
	return prog, p.err
}

type CommandLookuper interface {
	LookupCommand(name string) (*cmd.Command, bool, error)
}

type parser struct {
	l         *lexer
	curToken  token
	peekToken token
	cmdLookup CommandLookuper
	err       error
}

func newParser(l *lexer, cmdLookup CommandLookuper) *parser {
	p := &parser{
		l:         l,
		cmdLookup: cmdLookup,
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.nextToken()
}

func (p *parser) parse() *Program {
	prog := &Program{
		cmds: make([]*programCommand, 0),
	}

	for !p.isCurToken(tokenEOF) {
		cmd := p.parseCommand()
		if cmd != nil {
			prog.cmds = append(prog.cmds, cmd)
		}

		if p.err != nil {
			break
		}

		if p.isPeekToken(tokenEOF) {
			break
		}

		p.nextToken()
	}

	return prog
}

func (p *parser) isCurToken(t tokenType) bool {
	return p.curToken.typ == t
}

func (p *parser) isPeekToken(t tokenType) bool {
	return p.peekToken.typ == t
}

func (p *parser) tryPeek(t tokenType) bool {
	if p.isPeekToken(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *parser) expectPeek(t tokenType) bool {
	if !p.tryPeek(t) {
		p.peekError(t)
		return false
	}
	return true
}

func (p *parser) peekError(t tokenType) {
	p.error(fmt.Errorf("expected next token to be %s, got %s instead", t, p.peekToken.typ))
}

func (p *parser) errorf(format string, v ...any) {
	p.error(fmt.Errorf(format, v...))
}

func (p *parser) error(err error) {
	p.err = err
}

func (p *parser) parseCommand() *programCommand {
	pc := &programCommand{
		args: []argument{},
	}

	cmdName := ""
	switch p.curToken.typ {
	case tokenText:
		cmdName = p.parseText().literal
	case tokenSingleQuote:
		cmdName = p.parseSingleQuotes().literal
	case tokenDoubleQuote:
		cmdName = p.parseDoubleQuotes().literal
	default:
		return nil
	}

	cmd, found, err := p.cmdLookup.LookupCommand(cmdName)
	if err != nil {
		p.errorf("failed to lookup cmd: %s", err)
		return nil
	}
	if !found {
		p.error(fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound))
		return nil
	}
	pc.cmd = cmd

Loop:
	for {
		switch p.peekToken.typ {
		case tokenText, tokenSpace, tokenEscaped:
			p.nextToken()
			node := p.parseText()
			if len(node.literal) > 0 {
				pc.args = append(pc.args, node)
			}
		case tokenSingleQuote:
			p.nextToken()
			node := p.parseSingleQuotes()
			if len(node.literal) > 0 {
				pc.args = append(pc.args, node)
			}
		case tokenDoubleQuote:
			p.nextToken()
			node := p.parseDoubleQuotes()
			if len(node.literal) > 0 {
				pc.args = append(pc.args, node)
			}
		default:
			break Loop
		}
	}

	switch {
	case p.isPeekToken(tokenRedirect):
		p.parseRedirect(pc)
	case p.isPeekToken(tokenAppend):
		p.parseAppend(pc)
	}

	return pc
}

func (p *parser) parseRedirect(pc *programCommand) {
	assert.NotNil(pc)
	assert.Assert(p.isPeekToken(tokenRedirect))

	if !p.isCurToken(tokenSpace) {
		p.errorf("expected space before redirect token")
		return
	}

	p.nextToken()
	if !p.expectPeek(tokenSpace) {
		return
	}

	file := &fileRedirect{}

	switch {
	case strings.HasPrefix(p.curToken.literal, "1"):
		pc.stdOut = file
	case strings.HasPrefix(p.curToken.literal, "2"):
		pc.stdErr = file
	default:
		pc.stdOut = file
	}

	switch p.peekToken.typ {
	case tokenText, tokenSpace, tokenEscaped:
		if node := p.parseText(); node != nil {
			file.filename = node.literal
		}
	case tokenSingleQuote:
		if node := p.parseSingleQuotes(); node != nil {
			file.filename = node.literal
		}
	case tokenDoubleQuote:
		if node := p.parseDoubleQuotes(); node != nil {
			file.filename = node.literal
		}
	default:
		p.errorf("expected filename after redirect token but got %s", p.peekToken.typ)
	}
}

func (p *parser) parseAppend(pc *programCommand) {
	assert.NotNil(pc)
	assert.Assert(p.isPeekToken(tokenAppend))

	if !p.isCurToken(tokenSpace) {
		p.errorf("expected space before append token")
		return
	}

	p.nextToken()
	if !p.expectPeek(tokenSpace) {
		return
	}

	file := &fileAppend{}

	switch {
	case strings.HasPrefix(p.curToken.literal, "1"):
		pc.stdOut = file
	case strings.HasPrefix(p.curToken.literal, "2"):
		pc.stdErr = file
	default:
		pc.stdOut = file
	}

	switch p.peekToken.typ {
	case tokenText, tokenSpace, tokenEscaped:
		if node := p.parseText(); node != nil {
			file.filename = node.literal
		}
	case tokenSingleQuote:
		if node := p.parseSingleQuotes(); node != nil {
			file.filename = node.literal
		}
	case tokenDoubleQuote:
		if node := p.parseDoubleQuotes(); node != nil {
			file.filename = node.literal
		}
	default:
		p.errorf("expected filename after append token but got %s", p.peekToken.typ)
	}
}

func (p *parser) parseText() *rawText {
	str := ""
	for {
		switch p.curToken.typ {
		case tokenSpace:
			if len(str) > 0 {
				return &rawText{literal: str}
			}
			p.nextToken()
		case tokenText, tokenEscaped:
			str += p.curToken.literal
			p.nextToken()
		case tokenSingleQuote:
			if !p.isPeekToken(tokenSingleQuote) {
				return &rawText{literal: str}
			}
			p.nextToken()
			p.nextToken()
		case tokenDoubleQuote:
			if !p.isPeekToken(tokenDoubleQuote) {
				return &rawText{literal: str}
			}
			p.nextToken()
			p.nextToken()
		default:
			return &rawText{literal: str}
		}
	}
}

func (p *parser) parseSingleQuotes() *singleQuotedText {
	assert.Assert(p.isCurToken(tokenSingleQuote))

	node := &singleQuotedText{}
	builder := strings.Builder{}

Loop:
	for {
		switch {
		case p.isPeekToken(tokenText):
			p.nextToken()
			_, _ = builder.WriteString(p.curToken.literal)
		case p.isPeekToken(tokenSingleQuote):
			p.nextToken()
			if p.tryPeek(tokenSingleQuote) {
				_, _ = builder.WriteString(p.parseSingleQuotes().literal)
			}
			break Loop
		default:
			break Loop
		}
	}

	node.literal = builder.String()
	return node
}

func (p *parser) parseDoubleQuotes() *doubleQuotedText {
	assert.Assert(p.isCurToken(tokenDoubleQuote))

	node := &doubleQuotedText{}
	builder := strings.Builder{}

Loop:
	for {
		switch p.peekToken.typ {
		case tokenText, tokenEscaped:
			p.nextToken()
			_, _ = builder.WriteString(p.curToken.literal)
		case tokenDoubleQuote:
			p.nextToken()
			if p.tryPeek(tokenDoubleQuote) {
				_, _ = builder.WriteString(p.parseDoubleQuotes().literal)
			}

			if p.tryPeek(tokenText) {
				_, _ = builder.WriteString(p.parseText().literal)
			}
			break Loop
		default:
			break Loop
		}
	}

	node.literal = builder.String()
	return node
}
