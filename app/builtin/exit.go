package builtin

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func NewExitCommand(s *shell.Shell) *cmd.Command {
	return &cmd.Command{
		Name: "exit",
		Run: func(input []byte) error {
			return shell.ErrExit
		},
	}
}
