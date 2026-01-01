package shell

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewEchoCommand() *cmd.Command {
	return &cmd.Command{
		Name: "echo",
		Run: func(cmd *cmd.Command, args []string) error {
			assert.Assert(len(args) > 0)
			fmt.Fprintf(cmd.Stdout, "%s\n", strings.Join(args[1:], " "))
			return nil
		},
	}
}

func NewEchoCommandFunc() cmd.CommandFunc {
	return func() *cmd.Command {
		return NewEchoCommand()
	}
}
