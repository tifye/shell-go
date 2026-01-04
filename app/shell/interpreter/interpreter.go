package interpreter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/shell/interpreter/ast"
	"golang.org/x/sync/errgroup"
)

type (
	CmdFunc       func(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error
	CmdLookupFunc func(name string) (cmd CmdFunc, found bool, err error)
	EnvFunc       func(string) string
	OpenFileFunc  func(string, int, os.FileMode) (io.ReadWriteCloser, error)
)

type interpreterOption func(p *Interpreter)

func WithIO(stdin io.Reader, stdout, stderr io.Writer) interpreterOption {
	return func(p *Interpreter) {
		if stdin != nil {
			p.stdin = stdin
		}
		if stdout != nil {
			p.stdout = stdout
		}
		if stderr != nil {
			p.stderr = stderr
		}
	}
}

func WithCmdLookupFunc(f CmdLookupFunc) interpreterOption {
	return func(p *Interpreter) {
		if f != nil {
			p.cmdReg = f
		}
	}
}

func WithEnvFunc(f EnvFunc) interpreterOption {
	return func(p *Interpreter) {
		if f != nil {
			p.getenv = f
		}
	}
}

func WithOpenFileFunc(f OpenFileFunc) interpreterOption {
	return func(p *Interpreter) {
		if f != nil {
			p.openFile = f
		}
	}
}

type Interpreter struct {
	cmdReg   CmdLookupFunc
	getenv   EnvFunc
	openFile OpenFileFunc

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func DefaultInterpreter() *Interpreter {
	return &Interpreter{
		cmdReg:   func(name string) (cmd CmdFunc, found bool, err error) { return nil, false, nil },
		getenv:   func(s string) string { return "" },
		openFile: func(s string, i int, fm os.FileMode) (io.ReadWriteCloser, error) { return nil, os.ErrNotExist },
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

func NewInterpreter(opts ...interpreterOption) *Interpreter {
	p := DefaultInterpreter()
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Interpreter) Evaluate(input string) error {
	root, err := ast.Parse(input)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return p.eval(root)
}

func (p *Interpreter) eval(n ast.Node) error {
	switch n := n.(type) {
	case *ast.Root:
		return p.evalSequential(n.Cmds)
	case *ast.PipeStmt:
		return p.evalPipeline(n)
	case *ast.CommandStmt:
		return p.evalCmd(n, nil, nil)
	}
	return nil
}

func (p *Interpreter) evalSequential(stmts []ast.Statement) error {
	for _, stmt := range stmts {
		if err := p.eval(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *Interpreter) evalPipeline(pipe *ast.PipeStmt) error {
	if len(pipe.Cmds) == 0 {
		return nil
	}
	if len(pipe.Cmds) == 1 {
		return p.evalCmd(pipe.Cmds[0], nil, nil)
	}

	eg := errgroup.Group{}
	var pr *io.PipeReader
	for i := 0; i < len(pipe.Cmds)-1; i++ {
		nextReader, pw := io.Pipe()

		var r io.Reader
		if pr != nil {
			r = &ignoreClosedPipeRead{pr}
		}
		eg.Go(func() error {
			return p.evalCmd(pipe.Cmds[i], r, &ignoreClosedPipeWrite{pw})
		})

		pr = nextReader
	}

	eg.Go(func() error {
		return p.evalCmd(pipe.Cmds[len(pipe.Cmds)-1], &ignoreClosedPipeRead{pr}, nil)
	})

	return eg.Wait()
}

func (p *Interpreter) evalCmd(cmdStmt *ast.CommandStmt, r io.Reader, w io.Writer) error {
	cmdName, err := p.evalExpression(cmdStmt.Name)
	if err != nil {
		return fmt.Errorf("eval command name: %w", err)
	}

	cmdFunc, found, err := p.cmdReg(cmdName)
	if err != nil {
		return fmt.Errorf("look up command: %w", err)
	}
	if !found {
		return fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound)
	}

	args, err := p.evalArgsList(cmdStmt.Args)
	if err != nil {
		return fmt.Errorf("%s: eval args: %w", cmdName, err)
	}
	args = append([]string{cmdName}, args...)

	if r == nil {
		r = p.stdin
	} else {
		if c, ok := r.(io.Closer); ok {
			defer c.Close()
		}
	}

	stdouts := make([]io.Writer, 0)
	stderrs := make([]io.Writer, 0)

	if w != nil {
		stdouts = append(stdouts, w)
		if c, ok := w.(io.Closer); ok {
			defer c.Close()
		}
	} else {
		if len(cmdStmt.StdOut) == 0 {
			stdouts = append(stdouts, p.stdout)
		}
	}

	for _, n := range cmdStmt.StdOut {
		wr, err := p.evalStdOutStmt(n)
		if err != nil {
			return fmt.Errorf("%s: eval output writer: %w", cmdName, err)
		}
		if c, ok := wr.(io.Closer); ok {
			defer c.Close()
		}
		stdouts = append(stdouts, wr)
	}

	for _, n := range cmdStmt.StdErr {
		wr, err := p.evalStdOutStmt(n)
		if err != nil {
			return fmt.Errorf("%s: eval err output writer: %w", cmdName, err)
		}
		if c, ok := wr.(io.Closer); ok {
			defer c.Close()
		}
		stderrs = append(stderrs, wr)
	}

	stdout := io.MultiWriter(stdouts...)
	stderr := io.MultiWriter(stderrs...)

	return cmdFunc(context.TODO(), r, stdout, stderr, args)
}

func (p *Interpreter) evalStdOutStmt(stmt ast.Statement) (io.Writer, error) {
	switch n := stmt.(type) {
	case *ast.RedirectStmt:
		filename, err := p.evalExpression(n.Filename)
		if err != nil {
			return nil, fmt.Errorf("eval filename: %w", err)
		}

		return p.openFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	case *ast.AppendStmt:
		filename, err := p.evalExpression(n.Filename)
		if err != nil {
			return nil, fmt.Errorf("eval filename: %w", err)
		}

		return p.openFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	default:
		return nil, fmt.Errorf("unsupported stdout statement of type: %s", reflect.TypeOf(n).String())
	}
}

func (p *Interpreter) evalExpression(expr ast.Expression) (string, error) {
	switch n := expr.(type) {
	case *ast.RawTextExpr:
		return n.Literal, nil
	case *ast.SingleQuotedTextExpr:
		return n.Literal, nil
	case *ast.DoubleQuotedTextExpr:
		b := strings.Builder{}
		for _, e := range n.Expressions {
			s, err := p.evalExpression(e)
			if err != nil {
				return "", fmt.Errorf("eval double quoted expr: %w", err)
			}

			if _, err := b.WriteString(s); err != nil {
				return "", fmt.Errorf("failed to write to string builder: %s", err)
			}
		}
		return b.String(), nil
	case *ast.VariableExpr:
		return os.Expand(n.Literal, p.getenv), nil
	default:
		return "", fmt.Errorf("unsupported expression of type: %s", reflect.TypeOf(n).String())
	}
}

func (p *Interpreter) evalArgsList(argsList *ast.ArgsList) ([]string, error) {
	args := make([]string, 0, len(argsList.Args))
	for _, a := range argsList.Args {
		s, err := p.evalExpression(a)
		if err != nil {
			return nil, err
		}

		args = append(args, s)
	}

	return args, nil
}

type ignoreClosedPipeWrite struct {
	*io.PipeWriter
}
type ignoreClosedPipeRead struct {
	*io.PipeReader
}

func (p *ignoreClosedPipeWrite) Write(b []byte) (int, error) {
	n, err := p.PipeWriter.Write(b)
	if errors.Is(err, io.ErrClosedPipe) {
		return len(b), nil
	}
	return n, err
}

func (p *ignoreClosedPipeRead) Read(b []byte) (int, error) {
	n, err := p.PipeReader.Read(b)
	if errors.Is(err, io.ErrClosedPipe) {
		return len(b), io.EOF
	}
	return n, err
}
