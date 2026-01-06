package main

import (
	"fmt"
	"io"
	"os"
	"syscall/js"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
)

func main() {
	pr, pw := io.Pipe()
	jsshell := js.Global().Get("goshell")
	if jsshell.IsUndefined() {
		panic("goshell is undefined")
	}
	imports := jsshell.Get("imports")
	if imports.IsUndefined() {
		panic("goshell.imports is undefined")
	}
	imports.Set("write", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) <= 0 {
			return nil
		}

		input := args[0].String()
		_, _ = pw.Write([]byte(input))
		return nil
	}))

	exports := jsshell.Get("exports")

	s := &shell.Shell{
		Stdout: writeFunc(func(b []byte) (int, error) {
			_ = exports.Call("output", string(b))
			return len(b), nil
		}),
		Stderr: writeFunc(func(b []byte) (int, error) {
			_ = exports.Call("output", string(b))
			return len(b), nil
		}),
		Stdin:          pr,
		Env:            env{},
		FS:             filesystem{},
		HistoryContext: history.NewHistoryContext(history.NewInMemoryHistory()),
		ExecFunc: func(cmd *cmd.Command, path string, args []string) error {
			fmt.Println(args)
			return nil
		},
		FullPathFunc: func(s string) (string, error) {
			return s, nil
		},
	}
	if err := s.Run(); err != nil {
		panic(err)
	}
}

type writeFunc func([]byte) (int, error)

func (f writeFunc) Write(b []byte) (int, error) {
	return f(b)
}

type stdoutWriter struct{}

func (_ stdoutWriter) Write(b []byte) (int, error) {
	return fmt.Print(string(b))
}

type env struct{}

func (e env) Get(key string) string {
	return os.Getenv(key)
}
