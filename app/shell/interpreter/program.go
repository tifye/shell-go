package interpreter

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

type Program struct {
	cmds []*command
}

func (p *Program) Run() error {
	for _, c := range p.cmds {
		args, err := c.Args()
		if err != nil {
			return fmt.Errorf("failed to run %q command, got: %w", c.cmd.Name, err)
		}
		if err := c.cmd.Run(args); err != nil {
			return fmt.Errorf("failed to run %q command, got: %w", c.cmd.Name, err)
		}
	}
	return nil
}

type (
	command struct {
		cmd  *cmd.Command
		args []argument
	}

	rawText struct {
		literal string
	}

	singleQuotedText struct {
		literal string
	}
)

func (c command) Args() ([]string, error) {
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

type argument interface {
	String() (string, error)
}

func (t rawText) String() (string, error) {
	return t.literal, nil
}
func (t singleQuotedText) String() (string, error) {
	return t.literal, nil
}
