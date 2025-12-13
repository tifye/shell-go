package builtin

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func NewEchoCommand(s *shell.Shell) *cmd.Command {
	return &cmd.Command{
		Name: "echo",
		Run: func(args []string) error {
			fmt.Fprintf(s.Stdout, "%s\n", strings.Join(args[1:], " "))
			return nil
		},
	}
}
