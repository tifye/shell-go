package main

import (
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"golang.org/x/term"
)

func main() {
	run()
}

//go:noinline
func run() {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	fsys := gofs{}
	shell := &shell.Shell{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Stdin:      os.Stdin,
		Env:        goenv{},
		FS:         fsys,
		Exec:       goexec,
		HistoryCtx: history.NewHistoryContext(history.NewInMemoryHistory()),
		FullPath:   filepath.Abs,
	}
	shell.AddBuiltins(
		builtin.NewExitCommand(shell),
		builtin.NewEchoCommand(),
		builtin.NewTypeCommand(shell),
		builtin.NewHistoryCommand(shell.HistoryCtx, fsys),
	)
	if err := shell.Run(); err != nil {
		panic(err)
	}
}
