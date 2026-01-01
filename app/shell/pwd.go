package shell

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

func NewPWDCommandFunc(s *Shell) cmd.CommandFunc {
	return func() *cmd.Command {
		return &cmd.Command{
			Name: "pwd",
			Run: func(cmd *cmd.Command, args []string) error {
				_, err := fmt.Fprintln(cmd.Stdout, s.WorkingDir)
				return err
			},
		}
	}
}
