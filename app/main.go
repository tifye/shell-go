package main

import (
	"os"
	"path/filepath"

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

	hist := history.NewInMemoryHistory()
	histCtx := history.NewHistoryContext(hist)
	fsys := gofs{}
	s := &shell.Shell{
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
		Stdin:          os.Stdin,
		Env:            goenv{},
		FS:             fsys,
		Exec:           goexec,
		HistoryContext: histCtx,
		FullPath:       filepath.Abs,
	}

	// s.AddBuiltins(
	// 	shell.NewExitCommand(),
	// 	shell.NewEchoCommand(),
	// 	shell.NewTypeCommand(s, s.CommandRegistry),
	// 	shell.NewHistoryCommand(histCtx, fsys),
	// )
	if err := s.Run(); err != nil {
		panic(err)
	}
}
