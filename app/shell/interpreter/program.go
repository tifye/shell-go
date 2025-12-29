package interpreter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"golang.org/x/sync/errgroup"
)

type Program struct {
	cmds []*programCommand
}

func (p *Program) Run() error {
	eg := errgroup.Group{}
	eg.SetLimit(len(p.cmds))
	for _, c := range p.cmds {
		eg.Go(func() error {
			return runCommand(c)
		})
	}
	return eg.Wait()
}

func runCommand(pc *programCommand) error {
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
	programCommand struct {
		cmd    *cmd.Command
		stdOut commandOut
		stdErr commandOut
		stdIn  commandIn
		args   []argument
	}

	pipeInRedirect struct {
		pipeReader *io.PipeReader
	}
	// todo: implement a multi writer to allow for file and pipe redirect
	pipeOutRedirect struct {
		pipeWriter *io.PipeWriter
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

type commandIn interface {
	Reader() (io.Reader, error)
}

func (p *pipeInRedirect) Reader() (io.Reader, error) {
	return p.pipeReader, nil
}

type commandOut interface {
	Writer() (io.Writer, error)
}

func (p *pipeOutRedirect) Writer() (io.Writer, error) {
	return p.pipeWriter, nil
}
func (f *fileRedirect) Writer() (io.Writer, error) {
	dir := filepath.Dir(f.filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
}
func (f *fileAppend) Writer() (io.Writer, error) {
	dir := filepath.Dir(f.filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(f.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
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
