package shell

import (
	"os/exec"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

func newExecCommand(s *Shell, name, path string) *cmd.Command {
	return &cmd.Command{
		Name: name,
		Run: func(args []string) error {
			ecmd := &exec.Cmd{
				Path: path,
				Args: args,
			}
			ecmd.Stdin = s.Stdin
			ecmd.Stdout = s.Stdout
			ecmd.Stderr = s.Stdout
			return ecmd.Run()
		},
	}
}
