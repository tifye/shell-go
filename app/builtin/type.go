package builtin

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func NewTypeCommand(s *shell.Shell) *cmd.Command {
	return &cmd.Command{
		Name: "type",
		Run: func(args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected exactly one argument to the 'type' command")
			}

			cmdName := args[1]
			if _, found := s.LookupBuiltinCommand(cmdName); found {
				_, _ = fmt.Fprintf(s.Stdout, "%s is a shell builtin\n", cmdName)
				return nil
			}

			fmt.Fprintf(s.Stdout, "%s: not found\n", cmdName)
			return nil
		},
	}
}
