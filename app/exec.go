package main

import (
	"os/exec"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func goexec(c *cmd.Command, path string, args []string) error {
	assert.NotNil(c)
	assert.Assert(len(path) > 0)
	assert.Assert(len(args) > 0)

	cmd := &exec.Cmd{
		Path: path,
		Args: args,
	}
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	return cmd.Run()
}
