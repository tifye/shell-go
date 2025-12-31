package builtin

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewTypeCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name: "type",
		Run: func(cmd *cmd.Command, args []string) error {
			assert.Assert(len(args) > 0)

			if len(args) != 2 {
				return fmt.Errorf("expected exactly one argument to the 'type' command")
			}

			cmdName := args[1]
			if _, found := s.LookupBuiltinCommand(cmdName); found {
				_, _ = fmt.Fprintf(cmd.Stdout, "%s is a shell builtin\n", cmdName)
				return nil
			}

			path, _, found := s.LookupPathCommand(cmdName)
			if found {
				assert.Assert(len(path) > 0)
				_, _ = fmt.Fprintf(cmd.Stdout, "%s is %s\n", cmdName, path)
				return nil
			}

			fmt.Fprintf(cmd.Stdout, "%s: not found\n", cmdName)
			return nil
		},
	}
}
