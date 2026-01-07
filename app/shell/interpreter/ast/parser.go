package ast

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

type Parser struct {
	l         *lexer
	prevToken token
	curToken  token
	peekToken token
	err       error
}

func Parse(input string) (*Root, error) {
	l := newLexer(input)
	p := NewParser(l)
	return p.Parse(), p.err
}

func NewParser(l *lexer) *Parser {
	return &Parser{
		l: l,
	}
}

func (p *Parser) Parse() *Root {
	p.nextToken()
	p.nextToken()
	return &Root{
		Cmds: p.parseStatements(),
	}
}

func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.nextToken()
}

func (p *Parser) isPrevToken(t tokenType) bool {
	return p.prevToken.typ == t
}

func (p *Parser) isCurToken(t tokenType) bool {
	return p.curToken.typ == t
}

func (p *Parser) isPeekToken(t tokenType) bool {
	return p.peekToken.typ == t
}

func (p *Parser) tryPeek(t tokenType) bool {
	if p.isPeekToken(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) expectPeek(t tokenType) bool {
	if !p.tryPeek(t) {
		p.peekError(t)
		return false
	}
	return true
}

func (p *Parser) peekError(t tokenType) {
	p.error(fmt.Errorf("expected next token to be %s, got %s instead", t, p.peekToken.typ))
}

func (p *Parser) errorf(format string, v ...any) {
	p.error(fmt.Errorf(format, v...))
}

func (p *Parser) error(err error) {
	p.err = err
}

func (p *Parser) parseStatements() []Statement {
	stmts := make([]Statement, 0)

	for !p.isCurToken(tokenEOF) {
		cmd := p.parseCommand()
		if p.err != nil {
			break
		}

		switch p.curToken.typ {
		case tokenPipeline:
			stmts = append(stmts, p.parsePipline(cmd))
		default:
			stmts = append(stmts, cmd)
		}

		p.nextToken()
	}

	return stmts
}

func (p *Parser) parsePipline(first *CommandStmt) *PipeStmt {
	assert.Assert(p.isCurToken(tokenPipeline))

	pipe := &PipeStmt{
		Cmds: []*CommandStmt{first},
	}

	for p.isCurToken(tokenPipeline) {
		p.nextToken()

		cmd := p.parseCommand()
		if p.err != nil {
			break
		}

		if cmd.StdIn != nil {
			p.errorf("cannot redirect stdin for commands that are part of a pipeline")
			break
		}

		pipe.Cmds = append(pipe.Cmds, cmd)
	}

	return pipe
}

func (p *Parser) parseCommand() *CommandStmt {
	cmd := &CommandStmt{}

	for p.isCurToken(tokenSpace) {
		p.nextToken()
	}

	switch p.curToken.typ {
	case tokenVariable:
		cmd.Name = p.parseVariable()
	case tokenText, tokenEscaped:
		cmd.Name = p.parseText()
	case tokenSingleQuote:
		cmd.Name = p.parseSingleQuotes()
	case tokenDoubleQuote:
		cmd.Name = p.parseDoubleQuotes()
	default:
		p.errorf("unsupported token type for command name: %q", p.curToken.typ)
		return nil
	}

	cmd.Args = p.parseArgsList()

Loop:
	for {
		switch p.curToken.typ {
		case tokenRedirect:
			isStdout := !strings.HasPrefix(p.curToken.literal, "2")
			r := p.parseRedirect()
			if isStdout {
				cmd.StdOut = append(cmd.StdOut, r)
			} else {
				cmd.StdErr = append(cmd.StdOut, r)
			}
		case tokenAppend:
			isStdout := !strings.HasPrefix(p.curToken.literal, "2")
			a := p.parseAppend()
			if isStdout {
				cmd.StdOut = append(cmd.StdOut, a)
			} else {
				cmd.StdErr = append(cmd.StdOut, a)
			}
		default:
			break Loop
		}
	}

	return cmd
}

func (p *Parser) parseArgsList() (a *ArgsList) {
	for p.isCurToken(tokenSpace) {
		p.nextToken()
	}

	a = &ArgsList{
		Args: make([]Expression, 0),
	}
	for {
		switch p.curToken.typ {
		case tokenText, tokenSpace, tokenEscaped:
			if s := p.parseText(); s.Literal != "" {
				a.Args = append(a.Args, s)
			}
		case tokenSingleQuote:
			if s := p.parseSingleQuotes(); s.Literal != "" {
				a.Args = append(a.Args, s)
			}
		case tokenDoubleQuote:
			a.Args = append(a.Args, p.parseDoubleQuotes())
		case tokenVariable:
			a.Args = append(a.Args, p.parseVariable())
		default:
			return
		}
	}
}

func (p *Parser) parseVariable() *VariableExpr {
	assert.Assert(p.isCurToken(tokenVariable))

	v := &VariableExpr{
		Literal: p.curToken.literal,
	}

	p.nextToken()

	if len(v.Literal) <= 1 {
		p.errorf("empty variable expression")
	}

	return v
}

func (p *Parser) parseEscaped() string {
	assert.Assert(p.isCurToken(tokenEscaped))

	char := strings.TrimPrefix(p.curToken.literal, `\`)
	switch char {
	// case "t":
	// 	return "\t"
	// case "n":
	// 	return "\n"
	default:
		return char
	}
}

func (p *Parser) parseText() *RawTextExpr {
	str := ""
	for {
		switch p.curToken.typ {
		case tokenSpace:
			p.nextToken()
			if len(str) > 0 {
				return &RawTextExpr{Literal: str}
			}
		case tokenText:
			str += strings.TrimPrefix(p.curToken.literal, `\`)
			p.nextToken()
		case tokenEscaped:
			str += p.parseEscaped()
			p.nextToken()
		case tokenSingleQuote:
			if !p.isPeekToken(tokenSingleQuote) {
				return &RawTextExpr{Literal: str}
			}
			p.nextToken()
			p.nextToken()
		case tokenDoubleQuote:
			if !p.isPeekToken(tokenDoubleQuote) {
				return &RawTextExpr{Literal: str}
			}
			p.nextToken()
			p.nextToken()
		default:
			return &RawTextExpr{Literal: str}
		}
	}
}

func (p *Parser) parseSingleQuotes() *SingleQuotedTextExpr {
	assert.Assert(p.isCurToken(tokenSingleQuote))

	node := &SingleQuotedTextExpr{}
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

func (p *Parser) parseDoubleQuotes() *DoubleQuotedTextExpr {
	assert.Assert(p.isCurToken(tokenDoubleQuote))
	p.nextToken()

	node := &DoubleQuotedTextExpr{
		Expressions: make([]Expression, 0),
	}

Loop:
	for {
		switch p.curToken.typ {
		case tokenText, tokenEscaped:
			rt := p.parseText()
			node.Expressions = append(node.Expressions, rt)
		case tokenVariable:
			v := p.parseVariable()
			node.Expressions = append(node.Expressions, v)
		case tokenDoubleQuote:
			if p.tryPeek(tokenDoubleQuote) {
				dq := p.parseDoubleQuotes()
				node.Expressions = append(node.Expressions, dq)
			}

			if p.tryPeek(tokenText) {
				rt := p.parseText()
				node.Expressions = append(node.Expressions, rt)
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

func (p *Parser) parseRedirect() *RedirectStmt {
	assert.Assert(p.isCurToken(tokenRedirect))

	if !p.isPrevToken(tokenSpace) {
		p.errorf("expected space before redirect token")
		return nil
	}

	if !p.expectPeek(tokenSpace) {
		return nil
	}

	var filename Expression

	switch p.peekToken.typ {
	case tokenText, tokenSpace, tokenEscaped:
		filename = p.parseText()
	case tokenSingleQuote:
		filename = p.parseSingleQuotes()
	case tokenDoubleQuote:
		filename = p.parseDoubleQuotes()
	default:
		p.errorf("expected filename after redirect token but got %s", p.peekToken.typ)
	}

	return &RedirectStmt{Filename: filename}
}

func (p *Parser) parseAppend() *AppendStmt {
	assert.Assert(p.isCurToken(tokenAppend))

	if !p.isPrevToken(tokenSpace) {
		p.errorf("expected space before append token")
		return nil
	}

	if !p.expectPeek(tokenSpace) {
		return nil
	}

	var filename Expression

	switch p.peekToken.typ {
	case tokenText, tokenSpace, tokenEscaped:
		filename = p.parseText()
	case tokenSingleQuote:
		filename = p.parseSingleQuotes()
	case tokenDoubleQuote:
		filename = p.parseDoubleQuotes()
	default:
		p.errorf("expected filename after append token but got %s", p.peekToken.typ)
	}

	return &AppendStmt{Filename: filename}
}
