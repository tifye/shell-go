package interpreter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"golang.org/x/sync/errgroup"
)

type Program struct {
	cmds              []*Command
	CommandLookupFunc func(string) (*cmd.Command, bool, error)
}

func (p *Program) Run() error {
	eg := errgroup.Group{}
	eg.SetLimit(len(p.cmds))
	for _, c := range p.cmds {
		eg.Go(func() error {
			return p.runCommand(c)
		})
	}
	return eg.Wait()
}

func (p *Program) runCommand(pc *Command) error {
	cmdName, err := pc.Name.Expand()
	if err != nil {
		return fmt.Errorf("evaluating command name: %w", err)
	}
	if len(cmdName) == 0 {
		return fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound)
	}

	cmd, found, err := p.CommandLookupFunc(cmdName)
	if err != nil {
		return fmt.Errorf("lookuping cmd: %s", err)
	}
	if !found {
		return fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound)
	}

	pc.cmd = cmd

	if pc.stdIn != nil {
		stdInReader, err := pc.stdIn.OpenReader()
		if err != nil {
			return fmt.Errorf("failed to get stdin: %w", err)
		}
		if rc, ok := stdInReader.(io.Closer); ok {
			defer func() {
				_ = rc.Close()
			}()
		}
		pc.cmd.Stdin = stdInReader
	}

	if pc.stdOut != nil {
		stdOutWriter, err := pc.stdOut.OpenWriter()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %w", err)
		}
		if wc, ok := stdOutWriter.(io.Closer); ok {
			defer func() {
				_ = wc.Close()
			}()
		}
		pc.cmd.Stdout = stdOutWriter
	}

	if pc.stdErr != nil {
		stdErrWriter, err := pc.stdErr.OpenWriter()
		if err != nil {
			return fmt.Errorf("failed to get stderr: %w", err)
		}
		if wc, ok := stdErrWriter.(io.Closer); ok {
			defer func() {
				_ = wc.Close()
			}()
		}
		pc.cmd.Stderr = stdErrWriter
	}

	args, err := pc.ExpandArgs()
	if err != nil {
		return fmt.Errorf("failed to run %q command, got: %w", pc.cmd.Name, err)
	}

	if err := pc.cmd.Run(pc.cmd, args); err != nil {
		return fmt.Errorf("failed to run %q command, got: %w", pc.cmd.Name, err)
	}

	return nil
}

type Node interface {
	Pos() int
	End() int
}

type StringExpr interface {
	Node
	Expand() (string, error)
}

type InputSource interface {
	Node
	OpenReader() (io.Reader, error)
}

type OutputTarget interface {
	Node
	OpenWriter() (io.Writer, error)
}

type (
	Command struct {
		Name   StringExpr
		Args   *Arguments
		stdOut OutputTarget
		stdErr OutputTarget
		stdIn  InputSource
		cmd    *cmd.Command
	}

	Arguments struct {
		Args []StringExpr
	}

	PipeInRedirect struct {
		Pipe       int
		pipeReader *io.PipeReader
	}

	PipeOutRedirect struct {
		Pipe       int
		pipeWriter *io.PipeWriter
	}

	FileRedirect struct {
		Redirect int
		Filename StringExpr
	}

	FileAppend struct {
		Append   int
		Filename StringExpr
	}

	Variable struct {
		ValuePos   int
		Literal    string
		LookupFunc func(string) string
	}

	RawText struct {
		ValuePos int
		Literal  string
	}

	SingleQuotedText struct {
		ValuePos int
		Literal  string
	}

	DoubleQuotedText struct {
		StartQuote int
		EndQuote   int
		Nodes      []StringExpr
	}
)

func (x *Command) Pos() int { return x.Name.Pos() }
func (x *Arguments) Pos() int {
	if len(x.Args) == 0 {
		return 0
	}
	return x.Args[0].Pos()
}
func (x *PipeInRedirect) Pos() int   { return x.Pipe }
func (x *PipeOutRedirect) Pos() int  { return x.Pipe }
func (x *FileRedirect) Pos() int     { return x.Redirect }
func (x *FileAppend) Pos() int       { return x.Append }
func (x *Variable) Pos() int         { return x.ValuePos }
func (x *RawText) Pos() int          { return x.ValuePos }
func (x *SingleQuotedText) Pos() int { return x.ValuePos }
func (x *DoubleQuotedText) Pos() int { return x.StartQuote }

func (x *Command) End() int {
	switch {
	case x.stdIn != nil:
		return x.stdIn.End()
	case x.cmd.Stderr != nil:
		return x.stdErr.End()
	case x.cmd.Stderr != nil:
		return x.stdOut.End()
	case len(x.Args.Args) > 0:
		return x.Args.End()
	default:
		return x.Name.End()
	}
}
func (x *Arguments) End() int {
	if len(x.Args) == 0 {
		return 0
	}
	return x.Args[len(x.Args)-1].End()
}
func (x *PipeInRedirect) End() int   { return x.Pipe }
func (x *PipeOutRedirect) End() int  { return x.Pipe }
func (x *FileRedirect) End() int     { return x.Filename.End() }
func (x *FileAppend) End() int       { return x.Filename.End() }
func (x *Variable) End() int         { return x.ValuePos + utf8.RuneCountInString(x.Literal) }
func (x *RawText) End() int          { return x.ValuePos }
func (x *SingleQuotedText) End() int { return x.ValuePos }
func (x *DoubleQuotedText) End() int { return x.StartQuote }

func (v Variable) Expand() (string, error)         { return os.Expand(v.Literal, v.LookupFunc), nil }
func (t RawText) Expand() (string, error)          { return t.Literal, nil }
func (t SingleQuotedText) Expand() (string, error) { return t.Literal, nil }
func (t DoubleQuotedText) Expand() (string, error) {
	b := strings.Builder{}
	for i := range t.Nodes {
		s, err := t.Nodes[i].Expand()
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

func (x *Command) ExpandArgs() ([]string, error) {
	args, err := x.Args.Expand()
	if err != nil {
		return nil, fmt.Errorf("expanding args: %w", err)
	}
	args = append([]string{x.cmd.Name}, args...)
	return args, nil
}
func (x *Arguments) Expand() ([]string, error) {
	out := make([]string, 0, len(x.Args)+1)
	for i := range x.Args {
		arg, err := x.Args[i].Expand()
		if err != nil {
			return out, err
		}
		out = append(out, arg)
	}
	return out, nil
}

func (p *PipeInRedirect) OpenReader() (io.Reader, error) {
	return p.pipeReader, nil
}

func (p *PipeOutRedirect) OpenWriter() (io.Writer, error) {
	return p, nil
}
func (p *PipeOutRedirect) Write(b []byte) (int, error) {
	n, err := p.pipeWriter.Write(b)
	if errors.Is(err, io.ErrClosedPipe) {
		return len(b), nil
	}
	return n, err
}
func (p *PipeOutRedirect) Close() error {
	return p.pipeWriter.Close()
}

func (f *FileRedirect) OpenWriter() (io.Writer, error) {
	filename, err := f.Filename.Expand()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("making path: %w", err)
		}
	}
	return os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
}
func (f *FileAppend) OpenWriter() (io.Writer, error) {
	filename, err := f.Filename.Expand()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("making path: %w", err)
		}
	}
	return os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
