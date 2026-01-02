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
	cmdName, err := pc.Name.String()
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
		stdInReader, err := pc.stdIn.Reader()
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
		stdOutWriter, err := pc.stdOut.Writer()
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
		stdErrWriter, err := pc.stdErr.Writer()
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

	args, err := pc.Args()
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

type StringNode interface {
	Node
	String() (string, error)
}

type CommandIn interface {
	Node
	Reader() (io.Reader, error)
}

type CommandOut interface {
	Node
	Writer() (io.Writer, error)
}

type (
	Command struct {
		Name      StringNode
		Arguments []StringNode
		stdOut    CommandOut
		stdErr    CommandOut
		stdIn     CommandIn
		cmd       *cmd.Command
	}

	PipeInRedirect struct {
		Pipe       int
		pipeReader *io.PipeReader
	}
	// todo: implement a multi writer to allow for file and pipe redirect
	PipeOutRedirect struct {
		Pipe       int
		pipeWriter *io.PipeWriter
	}
	FileRedirect struct {
		Redirect int
		Filename StringNode
	}

	FileAppend struct {
		Append   int
		Filename StringNode
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
		Nodes      []StringNode
	}
)

func (x *Command) Pos() int          { return x.Name.Pos() }
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
	case len(x.Arguments) > 0:
		return x.Arguments[len(x.Arguments)-1].End()
	default:
		return x.Name.End()
	}
}
func (x *PipeInRedirect) End() int   { return x.Pipe }
func (x *PipeOutRedirect) End() int  { return x.Pipe }
func (x *FileRedirect) End() int     { return x.Filename.End() }
func (x *FileAppend) End() int       { return x.Filename.End() }
func (x *Variable) End() int         { return x.ValuePos + utf8.RuneCountInString(x.Literal) }
func (x *RawText) End() int          { return x.ValuePos }
func (x *SingleQuotedText) End() int { return x.ValuePos }
func (x *DoubleQuotedText) End() int { return x.StartQuote }

func (v Variable) String() (string, error)         { return os.Expand(v.Literal, v.LookupFunc), nil }
func (t RawText) String() (string, error)          { return t.Literal, nil }
func (t SingleQuotedText) String() (string, error) { return t.Literal, nil }
func (t DoubleQuotedText) String() (string, error) {
	b := strings.Builder{}
	for i := range t.Nodes {
		s, err := t.Nodes[i].String()
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

func (c Command) Args() ([]string, error) {
	out := make([]string, 0, len(c.Arguments)+1)
	out = append(out, c.cmd.Name)
	for i := range c.Arguments {
		arg, err := c.Arguments[i].String()
		if err != nil {
			return out, err
		}
		out = append(out, arg)
	}
	return out, nil
}

func (p *PipeInRedirect) Reader() (io.Reader, error) {
	return p.pipeReader, nil
}

func (p *PipeOutRedirect) Writer() (io.Writer, error) {
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

func (f *FileRedirect) Writer() (io.Writer, error) {
	filename, err := f.Filename.String()
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
func (f *FileAppend) Writer() (io.Writer, error) {
	filename, err := f.Filename.String()
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
