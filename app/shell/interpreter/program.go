package interpreter

import (
	"errors"
	"fmt"
	"io"
	"iter"
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
	var eg errgroup.Group
	for _, c := range p.cmds {
		if c.RunParallel {
			eg.Go(func() error {
				return p.runCommand(c)
			})
		} else {
			err := p.runCommand(c)
			err = errors.Join(err, eg.Wait())
			if err != nil {
				return err
			}
		}
	}
	return eg.Wait()
}

func (p *Program) runCommand(pc *Command) error {
	cmd, err := p.getCommand(pc)
	if err != nil {
		pc.Redirects.CloseAll()
		return err
	}

	pc.cmd = cmd

	if pc.Redirects != nil {
		if pc.Redirects.StdIn != nil {
			r, err := pc.Redirects.StdIn.OpenReader()
			if err != nil {
				return fmt.Errorf("failed to open stdin reader: %w", err)
			}
			defer r.Close()
			pc.cmd.Stdin = r
		}

		if len(pc.Redirects.StdOut) > 0 {
			writers := []io.Writer{}
			for i := range pc.Redirects.StdOut {
				w, err := pc.Redirects.StdOut[i].OpenWriter()
				if err != nil {
					return fmt.Errorf("failed to open stdout writer: %w", err)
				}
				defer w.Close()
				writers = append(writers, w)
			}

			mw := io.MultiWriter(writers...)
			pc.cmd.Stdout = mw
		}

		if len(pc.Redirects.StdErr) > 0 {
			writers := []io.Writer{}
			for i := range pc.Redirects.StdErr {
				w, err := pc.Redirects.StdErr[i].OpenWriter()
				if err != nil {
					return fmt.Errorf("failed to open stderr writer: %w", err)
				}
				defer w.Close()
				writers = append(writers, w)
			}

			mw := io.MultiWriter(writers...)
			pc.cmd.Stderr = mw
		}
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

func (p *Program) getCommand(pc *Command) (*cmd.Command, error) {
	cmdName, err := pc.Name.Expand()
	if err != nil {

		return nil, fmt.Errorf("evaluating command name: %w", err)
	}
	if len(cmdName) == 0 {

		return nil, fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound)
	}

	cmd, found, err := p.CommandLookupFunc(cmdName)
	if err != nil {

		return nil, fmt.Errorf("looking up cmd: %s", err)
	}
	if !found {

		return nil, fmt.Errorf("%s: %w", cmdName, ErrCommandNotFound)
	}

	return cmd, nil
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
	OpenReader() (io.ReadCloser, error)
}

type OutputTarget interface {
	Node
	OpenWriter() (io.WriteCloser, error)
}

type (
	Command struct {
		Name        StringExpr
		Args        *Arguments
		Redirects   *Redirects
		cmd         *cmd.Command
		RunParallel bool
	}

	Arguments struct {
		Args []StringExpr
	}

	Redirects struct {
		BeginPos int
		EndPos   int
		StdIn    InputSource
		StdOut   []OutputTarget
		StdErr   []OutputTarget
	}

	PipeInRedirect struct {
		Pipe       int
		PipeReader *io.PipeReader
	}

	PipeOutRedirect struct {
		Pipe       int
		PipeWriter *io.PipeWriter
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
func (x *Redirects) Pos() int        { return x.BeginPos }
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
	case x.Redirects != nil:
		return x.Redirects.End()
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
func (x *Redirects) End() int        { return x.EndPos }
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

func (p *PipeInRedirect) OpenReader() (io.ReadCloser, error) {
	return p.PipeReader, nil
}

func (p *PipeOutRedirect) OpenWriter() (io.WriteCloser, error) {
	return p, nil
}
func (p *PipeOutRedirect) Write(b []byte) (int, error) {
	n, err := p.PipeWriter.Write(b)
	if errors.Is(err, io.ErrClosedPipe) {
		return len(b), nil
	}
	return n, err
}
func (p *PipeOutRedirect) Close() error {
	return p.PipeWriter.Close()
}

func (f *FileRedirect) OpenReader() (io.ReadCloser, error) {
	filename, err := f.Filename.Expand()
	if err != nil {
		return nil, err
	}
	return os.OpenFile(filename, os.O_RDONLY, 0)
}
func (f *FileRedirect) OpenWriter() (io.WriteCloser, error) {
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
func (f *FileAppend) OpenWriter() (io.WriteCloser, error) {
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

// Instead of fixing my stupid implementation I
// saw an opportunity to try iterators for the
// first time
func (r *Redirects) Nodes() iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if r.StdIn != nil {
			if !yield(r.StdIn) {
				return
			}
		}

		for _, n := range r.StdOut {
			if !yield(n) {
				return
			}
		}

		for _, n := range r.StdErr {
			if !yield(n) {
				return
			}
		}
	}
}

// Nasty temp fix.
// What I need to do is build
// an actual interpreter and only let
// these be the syntax tree. Currently
// the program both defines the syntax tree
// and runs the commands which causes stupid
// things like this.
func (r *Redirects) CloseAll() {
	if r.StdIn != nil {
		if r, err := r.StdIn.OpenReader(); err == nil {
			r.Close()
		}
	}
	for i := range r.StdOut {
		if w, err := r.StdOut[i].OpenWriter(); err == nil {
			w.Close()
		}
	}

	for i := range r.StdErr {
		if w, err := r.StdOut[i].OpenWriter(); err == nil {
			w.Close()
		}
	}
}
