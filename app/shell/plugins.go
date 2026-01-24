package shell

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewPluginsCommandFunc(s *Shell) cmd.CommandFunc {
	assert.NotNil(s, "shell")
	return func() *cmd.Command {
		return &cmd.Command{
			Name: "plugins",
			Run: func(cmd *cmd.Command, args []string) error {
				assert.Assert(len(args) > 0)
				s.tr.Writer().StagePushForegroundColor(terminal.Rose)
				for i, p := range s.plugins {
					s.tr.Writer().Stagef("%d %s\n", i+1, p.Name())
				}
				s.tr.Writer().StagePopForegroundColor().
					Commit()
				return nil
			},
		}
	}
}
