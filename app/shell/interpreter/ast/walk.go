package ast

import "reflect"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func walkList[N Node](v Visitor, list []N) {
	for _, node := range list {
		Walk(v, node)
	}
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	if node == nil {
		return
	}

	switch n := node.(type) {
	case *Program:
		walkList(v, n.Chain)
	case *PipeStmt:
		walkList(v, n.Chain)
	case *ArgList:
		walkList(v, n.Args)
	case *CommandStmt:
		Walk(v, n.Name)
		Walk(v, n.Args)
		Walk(v, n.StdIn)
		walkList(v, n.StdOut)
		walkList(v, n.StdErr)
	case *RedirectStmt:
		Walk(v, n.Filename)
	case *AppendStmt:
		Walk(v, n.Filename)
	case *DoubleQuotedTextExpr:
		walkList(v, n.Expressions)
	case *VariableExpr, *SingleQuotedTextExpr, *RawTextExpr:
	default:
		panic("cannot walk node of type: " + reflect.TypeOf(n).String())
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
