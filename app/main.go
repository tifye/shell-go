package main

import (
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	shell := &shell.Shell{
		Stdout:   os.Stdout,
		Stdin:    os.Stdin,
		Env:      goenv{},
		FS:       gofs{},
		FullPath: filepath.Abs,
	}
	shell.AddBuiltins(
		builtin.NewExitCommand(shell),
		builtin.NewEchoCommand(shell),
		builtin.NewTypeCommand(shell),
	)
	if err := shell.Run(); err != nil {
		panic(err)
	}
}
