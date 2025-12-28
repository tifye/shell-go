package builtin

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewExitCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name:   "exit",
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		Stdin:  s.Stdin,
		Run: func(cmd *cmd.Command, args []string) error {
			return shell.ErrExit
		},
	}
}
