package main

import (
	"errors"
	"os/exec"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func goexec(c *cmd.Command, path string, args []string) error {
	assert.NotNil(c)
	assert.Assert(len(path) > 0)
	assert.Assert(len(args) > 0)

	cmd := &exec.Cmd{
		Path:   path,
		Args:   args,
		Stdin:  c.Stdin,
		Stdout: c.Stdout,
		Stderr: c.Stderr,
	}

	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}

	return nil
}
