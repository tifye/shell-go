package shell

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

func NewExitCommand() *cmd.Command {
	return &cmd.Command{
		Name: "exit",
		Run: func(cmd *cmd.Command, args []string) error {
			return ErrExit
		},
	}
}

func NewExitCommandFunc() cmd.CommandFunc {
	return func() *cmd.Command {
		return NewExitCommand()
	}
}
