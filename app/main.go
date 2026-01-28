package main

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/codecrafters-io/shell-starter-go/app/plugin"
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

	hctx := history.NewHistoryContext(history.NewInMemoryHistory())
	workingDir, _ := syscall.Getwd()

	s := &shell.Shell{
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
		Stdin:          os.Stdin,
		Env:            goenv{},
		FS:             gofs{},
		HistoryContext: hctx,
		ExecFunc:       goexec,
		FullPathFunc:   filepath.Abs,
		WorkingDir:     workingDir,
	}

	s.WithPlugins(
		plugin.NewAutoComplete(),
		plugin.NewNavHistory(),
		plugin.NewClearScreen(),
		plugin.ControlCExit{},
	)
	if os.Getenv("ENV") != "CODECRAFTERS" {
		s.WithPlugins(
			plugin.NewCompletionHints(),
			plugin.NewLuaPluginLoader(".plugins"),
		)
	}

	if err := s.Run(); err != nil {
		panic(err)
	}
}
