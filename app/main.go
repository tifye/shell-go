package main

import (
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
)

func main() {
	shell := &shell.Shell{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Stdin:    os.Stdin,
		Env:      goenv{},
		FS:       gofs{},
		Exec:     goexec,
		History:  history.NewInMemoryHistory(),
		FullPath: filepath.Abs,
	}
	shell.AddBuiltins(
		builtin.NewExitCommand(shell),
		builtin.NewEchoCommand(shell),
		builtin.NewTypeCommand(shell),
		builtin.NewHistoryCommand(shell),
	)
	if err := shell.Run(); err != nil {
		panic(err)
	}
}
