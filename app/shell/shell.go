package shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

var (
	// ErrExit causes the Shell to exit when returned by
	// a command
	ErrExit = errors.New("shell exited")
)

type Shell struct {
	Stdout   io.Writer
	Stdin    io.Reader
	builtins []*cmd.Command
}

func NewShell(w io.Writer, r io.Reader) *Shell {
	return &Shell{
		Stdout:   w,
		Stdin:    r,
		builtins: make([]*cmd.Command, 0),
	}
}

func (s *Shell) Run() error {
	assert.NotNil(s.Stdout)
	assert.NotNil(s.Stdin)

	reader := bufio.NewReader(s.Stdin)

	for {
		fmt.Fprint(s.Stdout, "$ ")
		input, err := reader.ReadBytes('\n')
		if err != nil {
			_, _ = fmt.Fprintf(s.Stdout, "error reading input: %s\n", err)
			return nil
		}

		input = bytes.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue
		}

		args := strings.Fields(string(input))
		cmdName := args[0]
		cmd, found := s.LookupCommand(cmdName)
		if !found {
			_, _ = fmt.Fprintf(s.Stdout, "%s: command not found\n", cmdName)
			continue
		}

		if err = cmd.Run(args); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}

			_, _ = fmt.Fprintf(s.Stdout, "error executing '%s': %s\n", input, err)
		}
	}
}

func (s *Shell) AddBuiltin(command *cmd.Command) {
	s.builtins = append(s.builtins, command)
}

func (s *Shell) LookupBuiltinCommand(name string) (*cmd.Command, bool) {
	for _, c := range s.builtins {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

func (s *Shell) LookupPathCommand(_ string) (*cmd.Command, bool) {
	return nil, false
}

func (s *Shell) LookupCommand(name string) (*cmd.Command, bool) {
	cmd, found := s.LookupBuiltinCommand(name)
	if found {
		assert.NotNil(cmd)
		return cmd, true
	}

	cmd, found = s.LookupPathCommand(name)
	if found {
		assert.NotNil(cmd)
		return cmd, true
	}

	return nil, false
}
