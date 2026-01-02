package interpreter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	cmdName, err := pc.name.String()
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

type (
	Command struct {
		name   StringNode
		cmd    *cmd.Command
		stdOut commandOut
		stdErr commandOut
		stdIn  commandIn
		args   []StringNode
	}

	PipeInRedirect struct {
		pipeReader *io.PipeReader
	}
	// todo: implement a multi writer to allow for file and pipe redirect
	PipeOutRedirect struct {
		pipeWriter *io.PipeWriter
	}
	FileRedirect struct {
		Filename string
	}
	FileAppend struct {
		Filename string
	}

	Variable struct {
		Literal    string
		LookupFunc func(string) string
	}
	RawText struct {
		Literal string
	}
	SingleQuotedText struct {
		Literal string
	}
	DoubleQuotedText struct {
		Nodes []StringNode
	}
)

func (c Command) Args() ([]string, error) {
	out := make([]string, 0, len(c.args)+1)
	out = append(out, c.cmd.Name)
	for i := range c.args {
		arg, err := c.args[i].String()
		if err != nil {
			return out, err
		}
		out = append(out, arg)
	}
	return out, nil
}

type commandIn interface {
	Reader() (io.Reader, error)
}

func (p *PipeInRedirect) Reader() (io.Reader, error) {
	return p.pipeReader, nil
}

type commandOut interface {
	Writer() (io.Writer, error)
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
	dir := filepath.Dir(f.Filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.Filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
}
func (f *FileAppend) Writer() (io.Writer, error) {
	dir := filepath.Dir(f.Filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

type StringNode interface {
	String() (string, error)
}

func (v Variable) String() (string, error) {
	return os.Expand(v.Literal, v.LookupFunc), nil
}
func (t RawText) String() (string, error) {
	return t.Literal, nil
}
func (t SingleQuotedText) String() (string, error) {
	return t.Literal, nil
}
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
