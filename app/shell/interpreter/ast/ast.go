package ast

import (
	"unicode/utf8"
)

type Root struct {
	Cmds []Statement
}

type Node interface {
	Pos() int
	End() int
}

type Statement interface {
	Node
	stmtNode()
}

type Expression interface {
	Node
	exprNode()
}

type (
	CommandStmt struct {
		Name   Expression
		Args   *ArgsList
		StdIn  Statement
		StdOut []Statement
		StdErr []Statement
	}

	PipeStmt struct {
		Cmds []*CommandStmt
	}

	BackgroundStmt struct {
		AmpersandPos int
		Stmt         Statement
	}

	ArgsList struct {
		Args []Expression
	}

	RedirectStmt struct {
		RedirectPos int
		Filename    Expression
	}

	AppendStmt struct {
		AppendPos int
		Filename  Expression
	}

	VariableExpr struct {
		ValuePos int
		Literal  string
	}

	RawTextExpr struct {
		ValuePos int
		Literal  string
	}

	SingleQuotedTextExpr struct {
		ValuePos int
		Literal  string
	}

	DoubleQuotedTextExpr struct {
		StartQuote  int
		EndQuote    int
		Expressions []Expression
	}
)

func (x *Root) Pos() int        { return x.Cmds[0].Pos() }
func (x *PipeStmt) Pos() int    { return x.Cmds[0].Pos() }
func (x *CommandStmt) Pos() int { return x.Name.Pos() }
func (x *ArgsList) Pos() int {
	if len(x.Args) == 0 {
		return 0
	}
	return x.Args[0].Pos()
}
func (x *RedirectStmt) Pos() int         { return x.RedirectPos }
func (x *AppendStmt) Pos() int           { return x.AppendPos }
func (x *VariableExpr) Pos() int         { return x.ValuePos }
func (x *RawTextExpr) Pos() int          { return x.ValuePos }
func (x *SingleQuotedTextExpr) Pos() int { return x.ValuePos }
func (x *DoubleQuotedTextExpr) Pos() int { return x.StartQuote }
func (x *BackgroundStmt) Pos() int       { return x.Stmt.Pos() }

func (x *Root) End() int     { return x.Cmds[0].End() }
func (x *PipeStmt) End() int { return x.Cmds[0].End() }
func (x *CommandStmt) End() int {
	switch {
	case len(x.Args.Args) > 0:
		return x.Args.End()
	default:
		return x.Name.End()
	}
}
func (x *ArgsList) End() int {
	if len(x.Args) == 0 {
		return 0
	}
	return x.Args[len(x.Args)-1].End()
}
func (x *RedirectStmt) End() int         { return x.Filename.End() }
func (x *AppendStmt) End() int           { return x.Filename.End() }
func (x *VariableExpr) End() int         { return x.ValuePos + utf8.RuneCountInString(x.Literal) }
func (x *RawTextExpr) End() int          { return x.ValuePos }
func (x *SingleQuotedTextExpr) End() int { return x.ValuePos }
func (x *DoubleQuotedTextExpr) End() int { return x.StartQuote }
func (x *BackgroundStmt) End() int       { return x.AmpersandPos }

func (*PipeStmt) stmtNode()       {}
func (*CommandStmt) stmtNode()    {}
func (*AppendStmt) stmtNode()     {}
func (*RedirectStmt) stmtNode()   {}
func (*BackgroundStmt) stmtNode() {}

func (*VariableExpr) exprNode()         {}
func (*RawTextExpr) exprNode()          {}
func (*SingleQuotedTextExpr) exprNode() {}
func (*DoubleQuotedTextExpr) exprNode() {}
