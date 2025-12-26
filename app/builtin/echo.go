package builtin

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewEchoCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name:   "echo",
		Stdout: s.Stdout,
		Stdin:  s.Stdin,
		Run: func(cmd *cmd.Command, args []string) error {
			assert.Assert(len(args) > 0)

			fmt.Fprintf(cmd.Stdout, "%s\n", strings.Join(args[1:], " "))
			return nil
		},
	}
}
