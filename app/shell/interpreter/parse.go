package interpreter

import (
	"errors"
	"fmt"
	"io"
	"math"
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
		cmds:              p.parseCommands(),
		CommandLookupFunc: p.cmdLookup.LookupCommand,
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

func (p *parser) parseCommands() []*Command {
	cmds := make([]*Command, 0)

	var pipeIn *PipeInRedirect
	for !p.isCurToken(tokenEOF) {
		cmd := p.parseCommand()
		if p.err != nil {
			break
		}
		assert.NotNil(cmd)

		if pipeIn != nil {
			if cmd.Redirects.StdIn != nil {
				// todo: no need to break here. Add support for multiple errors with source locations
				p.errorf("cannot have multiple command input redirects")
				break
			}

			cmd.Redirects.StdIn = pipeIn
		}

		cmds = append(cmds, cmd)

		if p.isCurToken(tokenPipeline) {
			pr, pw := io.Pipe()
			cmd.Redirects.StdOut = append(cmd.Redirects.StdOut, &PipeOutRedirect{p.curToken.pos, pw})
			pipeIn = &PipeInRedirect{p.curToken.pos, pr}
		}

		p.nextToken()
	}

	return cmds
}

func (p *parser) parseCommand() (pc *Command) {
	pc = &Command{}

	for p.isCurToken(tokenSpace) {
		p.nextToken()
	}

	switch p.curToken.typ {
	case tokenVariable:
		pc.Name = p.parseVariable()
	case tokenText, tokenEscaped:
		pc.Name = p.parseText()
	case tokenSingleQuote:
		pc.Name = p.parseSingleQuotes()
	case tokenDoubleQuote:
		pc.Name = p.parseDoubleQuotes()
	default:
		p.errorf("unsupported token type for command name: %q", p.curToken.typ)
		return nil
	}

	pc.Args = p.parseArguments()
	pc.Redirects = p.parseRedirects()

	return pc
}

func (p *parser) parseArguments() (a *Arguments) {
	a = &Arguments{
		Args: make([]StringExpr, 0),
	}
	for {
		switch p.curToken.typ {
		case tokenText, tokenSpace, tokenEscaped:
			node := p.parseText()
			if node != nil && len(node.Literal) > 0 {
				a.Args = append(a.Args, node)
			}
		case tokenSingleQuote:
			node := p.parseSingleQuotes()
			if node != nil && len(node.Literal) > 0 {
				a.Args = append(a.Args, node)
			}
		case tokenDoubleQuote:
			node := p.parseDoubleQuotes()
			if node != nil && len(node.Nodes) > 0 {
				a.Args = append(a.Args, node)
			}
		case tokenVariable:
			node := p.parseVariable()
			if node != nil {
				a.Args = append(a.Args, node)
			}
		default:
			return
		}
	}
}

func (p *parser) parseRedirects() *Redirects {
	r := &Redirects{
		BeginPos: math.MaxInt,
		EndPos:   math.MinInt,
		StdOut:   make([]OutputTarget, 0),
		StdErr:   make([]OutputTarget, 0),
	}

Loop:
	for {
		switch p.curToken.typ {
		case tokenRedirect, tokenAppend:
			p.parseRedirect(r)
		default:
			break Loop
		}
	}

	for n := range r.Nodes() {
		if n.Pos() < r.BeginPos {
			r.BeginPos = n.Pos()
		}
		if n.End() > r.EndPos {
			r.BeginPos = n.End()
		}
	}

	return r
}

func (p *parser) parseVariable() *Variable {
	assert.Assert(p.isCurToken(tokenVariable))

	v := &Variable{
		Literal:    p.curToken.literal,
		LookupFunc: p.getEnvFunc,
	}

	p.nextToken()

	if len(v.Literal) <= 1 {
		p.errorf("inavlid variable usage")
	}

	return v
}

func (p *parser) parseRedirect(r *Redirects) {
	assert.Assert(p.isCurToken(tokenAppend) || p.isCurToken(tokenRedirect))

	if !p.isPrevToken(tokenSpace) {
		p.errorf("expected space before redirect token")
		return
	}

	isRedirect := !p.isCurToken(tokenAppend)
	isStdout := !strings.HasPrefix(p.curToken.literal, "2")

	if !p.expectPeek(tokenSpace) {
		return
	}

	var filename StringExpr

	switch p.peekToken.typ {
	case tokenText, tokenSpace, tokenEscaped:
		if node := p.parseText(); node != nil {
			filename = node
		}
	case tokenSingleQuote:
		if node := p.parseSingleQuotes(); node != nil {
			filename = node
		}
	case tokenDoubleQuote:
		if node := p.parseDoubleQuotes(); node != nil {
			filename = node
		}
	default:
		p.errorf("expected filename after redirect token but got %s", p.peekToken.typ)
	}

	var out OutputTarget
	if isRedirect {
		out = &FileRedirect{Filename: filename}
	} else {
		out = &FileAppend{Filename: filename}
	}

	if isStdout {
		r.StdOut = append(r.StdOut, out)
	} else {
		r.StdErr = append(r.StdErr, out)
	}
}

func (p *parser) parseText() *RawText {
	str := ""
	for {
		switch p.curToken.typ {
		case tokenSpace:
			p.nextToken()
			if len(str) > 0 {
				return &RawText{Literal: str}
			}
		case tokenText, tokenEscaped:
			str += strings.TrimPrefix(p.curToken.literal, `\`)
			p.nextToken()
		case tokenSingleQuote:
			if !p.isPeekToken(tokenSingleQuote) {
				return &RawText{Literal: str}
			}
			p.nextToken()
			p.nextToken()
		case tokenDoubleQuote:
			if !p.isPeekToken(tokenDoubleQuote) {
				return &RawText{Literal: str}
			}
			p.nextToken()
			p.nextToken()
		default:
			return &RawText{Literal: str}
		}
	}
}

func (p *parser) parseSingleQuotes() *SingleQuotedText {
	assert.Assert(p.isCurToken(tokenSingleQuote))

	node := &SingleQuotedText{}
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
				_, _ = builder.WriteString(p.parseSingleQuotes().Literal)
			}
			p.nextToken()
			break Loop
		default:
			break Loop
		}
	}

	node.Literal = builder.String()
	return node
}

func (p *parser) parseDoubleQuotes() *DoubleQuotedText {
	assert.Assert(p.isCurToken(tokenDoubleQuote))
	p.nextToken()

	node := &DoubleQuotedText{
		Nodes: make([]StringExpr, 0),
	}

Loop:
	for {
		switch p.curToken.typ {
		case tokenText, tokenEscaped:
			rt := p.parseText()
			node.Nodes = append(node.Nodes, rt)
		case tokenVariable:
			v := p.parseVariable()
			node.Nodes = append(node.Nodes, v)
		case tokenDoubleQuote:
			if p.tryPeek(tokenDoubleQuote) {
				dq := p.parseDoubleQuotes()
				node.Nodes = append(node.Nodes, dq)
			}

			if p.tryPeek(tokenText) {
				rt := p.parseText()
				node.Nodes = append(node.Nodes, rt)
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
