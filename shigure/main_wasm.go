package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"syscall/js"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	pr, pw := io.Pipe()
	js.Global().Set("write", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) <= 0 {
			return nil
		}

		input := args[0].String()
		_, _ = pw.Write([]byte(input))
		return nil
	}))

	s := &shell.Shell{
		Stdout: writeFunc(func(b []byte) (int, error) {
			_ = js.Global().Call("output", string(b))
			return len(b), nil
		}),
		Stdin:    pr,
		Env:      env{},
		FS:       filesystem{},
		Exec:     exec,
		FullPath: fullpath,
	}
	s.AddBuiltins(
		builtin.NewExitCommand(s),
		builtin.NewEchoCommand(s),
		builtin.NewTypeCommand(s),
	)
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
	return ""
}

type filesystem struct{}

func (_ filesystem) Open(name string) (fs.File, error) {
	return nil, errors.New("not implemented")
}

func (_ filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

func exec(s *shell.Shell, path string, args []string) error {
	fmt.Println(path)
	return nil
}

func fullpath(path string) (string, error) {
	return path, nil
}
