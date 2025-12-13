package main

import (
	"os"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	shell := shell.NewShell(os.Stdout, os.Stdin)
	shell.AddBuiltin(builtin.NewExitCommand(shell))
	if err := shell.Run(); err != nil {
		panic(err)
	}
}
