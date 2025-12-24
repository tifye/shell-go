package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func Parse(input string, cmdLookup CommandLookuper) (*Program, error) {
	l := newLexer(input)
	p := newParser(l, cmdLookup)
	prog := p.parse()
	if len(p.errors) > 0 {
		err := errors.Join(p.errors...)
		return nil, fmt.Errorf("failed to parse with one or more errors: %w", err)
	}
	return prog, nil
}

type CommandLookuper interface {
	LookupCommand(name string) (*cmd.Command, bool, error)
}

type parser struct {
	l         *lexer
	curToken  token
	peekToken token
	cmdLookup CommandLookuper
	errors    []error
}

func newParser(l *lexer, cmdLookup CommandLookuper) *parser {
	p := &parser{
		l:         l,
		errors:    []error{},
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
		cmds: make([]*command, 0),
	}

	for !p.isCurToken(tokenEOF) {
		cmd := p.parseCommand()
		if cmd != nil {
			prog.cmds = append(prog.cmds, cmd)
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
	err := fmt.Errorf("expected next token to be %d, got %d instead", t, p.peekToken.typ)
	p.errors = append(p.errors, err)
}

func (p *parser) errorf(format string, v ...any) {
	err := fmt.Errorf(format, v...)
	p.errors = append(p.errors, err)
}

func (p *parser) parseCommand() *command {
	pc := &command{
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
		p.errorf("could not find cmd %s", cmdName)
		return nil
	}
	pc.cmd = cmd

	for {
		switch p.peekToken.typ {
		case tokenText, tokenSpace:
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
			return pc
		}
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
		case tokenText:
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
		switch {
		case p.isPeekToken(tokenText):
			p.nextToken()
			_, _ = builder.WriteString(p.curToken.literal)
		case p.isPeekToken(tokenDoubleQuote):
			p.nextToken()
			if p.tryPeek(tokenDoubleQuote) {
				_, _ = builder.WriteString(p.parseDoubleQuotes().literal)
			}
			break Loop
		default:
			break Loop
		}
	}

	node.literal = builder.String()
	return node
}
