package builtin

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewHistoryCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name:   "history",
		Stdin:  s.Stdin,
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		Run: func(cmd *cmd.Command, args []string) error {
			return nil
		},
	}
}
