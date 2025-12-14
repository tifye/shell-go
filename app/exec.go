package main

import (
	"os/exec"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func goexec(s *shell.Shell, path string, args []string) error {
	assert.NotNil(s)
	assert.Assert(len(path) > 0)
	assert.Assert(len(args) > 0)

	cmd := &exec.Cmd{
		Path: path,
		Args: args,
	}
	cmd.Stdin = s.Stdin
	cmd.Stdout = s.Stdout
	cmd.Stderr = s.Stdout
	return cmd.Run()
}
