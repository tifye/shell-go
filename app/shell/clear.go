package shell

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

func NewClearCommandFunc() cmd.CommandFunc {
	return func() *cmd.Command {
		return &cmd.Command{
			Name: "type",
			Run: func(cmd *cmd.Command, args []string) error {
				_, _ = cmd.Stdout.Write([]byte{0x1B, '[', '2', 'J'})
				return nil
			},
		}
	}
}
