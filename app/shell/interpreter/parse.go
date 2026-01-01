package interpreter

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

var (
	ErrCommandNotFound = errors.New("command not found")
)

func Parse(input string, cmdLookup registry, getEnv getEnvFunc) (*Program, error) {
	l := newLexer(input)
	p := newParser(l, cmdLookup, getEnv)
	prog := p.parse()
	return prog, p.err
}

type registry interface {
	LookupCommand(name string) (*cmd.Command, bool, error)
}

type getEnvFunc func(string) string

type parser struct {
	l          *lexer
	prevToken  token
	curToken   token
	peekToken  token
	cmdLookup  registry
	getEnvFunc getEnvFunc
	err        error
}

func newParser(l *lexer, cmdLookup registry, getEnv getEnvFunc) *parser {
	p := &parser{
		l:          l,
		cmdLookup:  cmdLookup,
		getEnvFunc: getEnv,
	}
	return p
}

func (p *parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.nextToken()
}

func (p *parser) parse() *Program {
	p.nextToken()
	p.nextToken()
	prog := &Program{
		cmds: p.parseCommands(),
	}
	return prog
}

func (p *parser) isPrevToken(t tokenType) bool {
	return p.prevToken.typ == t
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

func (p *parser) parseCommands() []*programCommand {
	cmds := make([]*programCommand, 0)

	var pipeIn *pipeInRedirect
	for !p.isCurToken(tokenEOF) {
		cmd := p.parseCommand()
		if p.err != nil {
			break
		}
		assert.NotNil(cmd)

		if pipeIn != nil {
			cmd.stdIn = pipeIn
			pipeIn = nil
		}

		cmds = append(cmds, cmd)

		if p.isCurToken(tokenPipeline) {
			pr, pw := io.Pipe()
			cmd.stdOut = &pipeOutRedirect{pw}
			pipeIn = &pipeInRedirect{pr}
		}

		p.nextToken()
	}

	return cmds
}

func (p *parser) parseCommand() *programCommand {
	pc := &programCommand{
		args: []StringNode{},
	}

	for p.isCurToken(tokenSpace) {
		p.nextToken()
	}

	cmdName := ""
	switch p.curToken.typ {
	case tokenText:
		cmdName = p.parseText().literal
	case tokenSingleQuote:
		cmdName = p.parseSingleQuotes().literal
	case tokenDoubleQuote:
		cmdName, _ = p.parseDoubleQuotes().String()
	default:
		p.errorf("unsupported token type for command name: %q", p.curToken.typ)
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
		switch p.curToken.typ {
		case tokenText, tokenSpace, tokenEscaped:
			node := p.parseText()
			if node != nil && len(node.literal) > 0 {
				pc.args = append(pc.args, node)
			}
		case tokenSingleQuote:
			node := p.parseSingleQuotes()
			if node != nil && len(node.literal) > 0 {
				pc.args = append(pc.args, node)
			}
		case tokenDoubleQuote:
			node := p.parseDoubleQuotes()
			if node != nil && len(node.parts) > 0 {
				pc.args = append(pc.args, node)
			}
		case tokenVariable:
			node := p.parseVariable()
			if node != nil {
				pc.args = append(pc.args, node)
			}
		case tokenRedirect:
			p.parseRedirect(pc)
			break Loop
		case tokenAppend:
			p.parseAppend(pc)
			break Loop
		default:
			break Loop
		}
	}

	return pc
}

func (p *parser) parseVariable() *variable {
	assert.Assert(p.isCurToken(tokenVariable))

	v := &variable{
		literal: p.curToken.literal,
		lookup:  p.getEnvFunc,
	}

	p.nextToken()

	if len(v.literal) <= 1 {
		p.errorf("inavlid variable usage")
	}

	return v
}

func (p *parser) parseRedirect(pc *programCommand) {
	assert.NotNil(pc)
	assert.Assert(p.isCurToken(tokenRedirect))

	if !p.isPrevToken(tokenSpace) {
		p.errorf("expected space before redirect token")
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

	if !p.expectPeek(tokenSpace) {
		return
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
			file.filename, _ = node.String()
		}
	default:
		p.errorf("expected filename after redirect token but got %s", p.peekToken.typ)
	}
}

func (p *parser) parseAppend(pc *programCommand) {
	assert.NotNil(pc)
	assert.Assert(p.isCurToken(tokenAppend))

	if !p.isPrevToken(tokenSpace) {
		p.errorf("expected space before append token")
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

	if !p.expectPeek(tokenSpace) {
		return
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
			file.filename, _ = node.String()
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
			p.nextToken()
			if len(str) > 0 {
				return &rawText{literal: str}
			}
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
			p.nextToken()
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
	p.nextToken()

	node := &doubleQuotedText{
		parts: make([]StringNode, 0),
	}

Loop:
	for {
		switch p.curToken.typ {
		case tokenText, tokenEscaped:
			rt := p.parseText()
			node.parts = append(node.parts, rt)
		case tokenVariable:
			v := p.parseVariable()
			node.parts = append(node.parts, v)
		case tokenDoubleQuote:
			if p.tryPeek(tokenDoubleQuote) {
				dq := p.parseDoubleQuotes()
				node.parts = append(node.parts, dq)
			}

			if p.tryPeek(tokenText) {
				rt := p.parseText()
				node.parts = append(node.parts, rt)
			} else {
				p.nextToken()
			}
			break Loop
		default:
			break Loop
		}
	}

	return node
}
