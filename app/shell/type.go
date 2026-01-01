package shell

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewTypeCommand(registry *cmd.Registry) *cmd.Command {
	assert.NotNil(registry, "registry")
	return &cmd.Command{
		Name: "type",
		Run: func(cmd *cmd.Command, args []string) error {
			assert.Assert(len(args) > 0)

			if len(args) != 2 {
				return fmt.Errorf("expected exactly one argument to the 'type' command")
			}

			cmdName := args[1]
			if _, found := registry.LookupBuiltinCommand(cmdName); found {
				_, _ = fmt.Fprintf(cmd.Stdout, "%s is a shell builtin\n", cmdName)
				return nil
			}

			path, _, found := registry.LookupPathCommand(cmdName)
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

func NewTypeCommandFunc(r *cmd.Registry) cmd.CommandFunc {
	return func() *cmd.Command {
		return NewTypeCommand(r)
	}
}
