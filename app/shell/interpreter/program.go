package interpreter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

type Program struct {
	cmds []*programCommand
}

func (p *Program) Run() error {
	for _, c := range p.cmds {
		if err := runCommand(c); err != nil {
			return err
		}
	}
	return nil
}

func runCommand(pc *programCommand) error {
	if pc.stdOut != nil {
		stdOutWriter, err := pc.stdOut.Writer()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %w", err)
		}
		defer func() {
			_ = stdOutWriter.Close()
		}()
		pc.cmd.Stdout = stdOutWriter
	}

	if pc.stdErr != nil {
		stdErrWriter, err := pc.stdOut.Writer()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %w", err)
		}
		defer func() {
			_ = stdErrWriter.Close()
		}()
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
	programCommand struct {
		cmd    *cmd.Command
		stdOut commandOut
		stdErr commandOut
		args   []argument
	}

	fileRedirect struct {
		filename string
	}
	fileAppend struct {
		filename string
	}

	rawText struct {
		literal string
	}
	singleQuotedText struct {
		literal string
	}
	doubleQuotedText struct {
		literal string
	}
)

func (c programCommand) Args() ([]string, error) {
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

type commandOut interface {
	Writer() (io.WriteCloser, error)
}

func (f *fileRedirect) Writer() (io.WriteCloser, error) {
	dir := filepath.Dir(f.filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.filename, os.O_TRUNC|os.O_CREATE, 0644)
}
func (f *fileAppend) Writer() (io.WriteCloser, error) {
	dir := filepath.Dir(f.filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.filename, os.O_APPEND|os.O_CREATE, 0644)
}

type argument interface {
	String() (string, error)
}

func (t rawText) String() (string, error) {
	return t.literal, nil
}
func (t singleQuotedText) String() (string, error) {
	return t.literal, nil
}
func (t doubleQuotedText) String() (string, error) {
	return t.literal, nil
}
