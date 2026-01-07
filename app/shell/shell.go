package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/codecrafters-io/shell-starter-go/app/shell/interpreter"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

var (
	// ErrExit causes the Shell to exit when returned by
	// a command
	ErrExit = errors.New("shell exited")
)

type env interface {
	Get(string) string
}

type FS interface {
	fs.ReadDirFS
	OpenFile(string, int) (io.ReadWriteCloser, error)
}

type Shell struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	Env          env
	FS           FS
	ExecFunc     func(cmd *cmd.Command, path string, args []string) error
	FullPathFunc func(string) (string, error)

	WorkingDir string

	HistoryContext  *history.HistoryContext
	CommandRegistry *cmd.Registry

	interp        *interpreter.Interpreter
	autocompleter *autocompleter
	tr            *terminal.TermReader
	tw            *terminal.TermWriter
}

func (s *Shell) buildPathCommandFunc(exec, path string) cmd.CommandFunc {
	return func() *cmd.Command {
		return &cmd.Command{
			Name:   exec,
			Stdout: s.Stdout,
			Stderr: s.Stderr,
			Stdin:  s.Stdin,
			Run: func(cmd *cmd.Command, args []string) error {
				return s.ExecFunc(cmd, path, args)
			},
		}
	}
}

func (s *Shell) Run() error {
	assert.NotNil(s.Stdout)
	assert.NotNil(s.Stderr)
	assert.NotNil(s.Stdin)

	s.tw = terminal.NewTermWriter(s.Stdout)
	s.tr = terminal.NewTermReader(s.Stdin, s.tw)
	s.Stdout = s.tw
	s.Stderr = s.tw

	if s.CommandRegistry == nil {
		registry, err := cmd.LoadFromPathEnv(s.Env.Get("PATH"), s.FS, s.FullPathFunc, s.buildPathCommandFunc)
		if err != nil {
			fmt.Fprintf(s.Stderr, "failed to load commands from PATH; %s\n", err)
			registry = cmd.NewResitry(s.buildPathCommandFunc)
		}

		registry.AddBuiltinCommand("type", NewTypeCommandFunc(registry))
		registry.AddBuiltinCommand("echo", NewEchoCommandFunc())
		registry.AddBuiltinCommand("history", NewHistoryCommandFunc(s.HistoryContext, s.FS))
		registry.AddBuiltinCommand("exit", NewExitCommandFunc())
		registry.AddBuiltinCommand("pwd", NewPWDCommandFunc(s))
		registry.AddBuiltinCommand("cd", NewCDCommandFunc(s))
		registry.AddBuiltinCommand("clear", NewClearCommandFunc())

		s.CommandRegistry = registry
	}

	s.autocompleter = &autocompleter{
		registry: s.CommandRegistry,
		RingTheBell: func() {
			s.tw.Write([]byte{0x07})
		},
		PossibleCompletions: func(p []string) {
			fmt.Fprintf(s.Stdout, "\n%s\n", strings.Join(p, "  "))
			fmt.Fprintf(s.Stdout, "$ %s", s.tr.Line())
		},
	}

	s.interp = interpreter.NewInterpreter(
		interpreter.WithIO(s.Stdin, s.Stdout, s.Stderr),
		interpreter.WithEnvFunc(s.Env.Get),
		interpreter.WithCmdLookupFunc(s.LookupCommand),
		interpreter.WithOpenFileFunc(func(name string, flags int, fm os.FileMode) (io.ReadWriteCloser, error) {
			return s.FS.OpenFile(name, flags)
		}),
	)

	if histFile := s.Env.Get("HISTFILE"); len(histFile) > 0 {
		err := history.ReadHistoryFromFile(s.HistoryContext, s.FS, s.Env.Get("HISTFILE"))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(s.Stderr, "failed to load command history: %s\n", err)
			}
		}
	}
	defer func() {
		histFile := s.Env.Get("HISTFILE")
		if len(histFile) <= 0 {
			return
		}
		err := history.AppendHistoryToFile(s.HistoryContext, s.FS, s.Env.Get("HISTFILE"))
		if err != nil {
			fmt.Fprintf(s.Stderr, "failed to save command history: %s\n", err)
		}
	}()

	s.repl()
	return nil
}

func (s *Shell) repl() {
	for {
		fmt.Fprint(s.Stdout, "$ ")

		input, err := s.read()
		if err != nil {
			if errors.Is(err, ErrExit) {
				return
			}
			_, _ = fmt.Fprintf(s.Stdout, "error reading input: %s\n", err)
			return
		}
		input = strings.TrimPrefix(input, "$ ")

		s.HistoryContext.Add(input)

		if err = s.interp.Evaluate(input); err != nil {
			if errors.Is(err, ErrExit) {
				return
			}

			if errors.Is(err, interpreter.ErrCommandNotFound) {
				_, _ = fmt.Fprintln(s.Stdout, err)
			} else {
				_, _ = fmt.Fprintf(s.Stdout, "error: %s\n", err)
			}
			continue
		}
	}
}

func (s *Shell) read() (string, error) {
	// Not exactly optimal but works for now
	histNavCtx := history.NewHistoryContext(s.HistoryContext.History)

	for {
		switch item := s.tr.NextItem(); item.Type {
		case terminal.ItemKeyUp:
			if item, ok := histNavCtx.Back(); ok {
				s.tr.ReplaceWith("$ " + item)
			}
		case terminal.ItemKeyDown:
			if item, ok := histNavCtx.Forward(); ok {
				s.tr.ReplaceWith("$ " + item)
			}
		case terminal.ItemKeyCtrlC:
			return "", ErrExit
		case terminal.ItemLineInput:
			return item.Literal, nil
		case terminal.ItemKeyTab:
			line, ok := s.autocompleter.Complete(s.tr.Line())
			if ok {
				s.tr.ReplaceWith("$ " + line)
			}
		case terminal.ItemKeyCtrlL:
			line := s.tr.Line()
			s.tw.Stage([]byte{0x1b, '[', '2', 'J'}) // clear terminal
			s.tw.Stage([]byte{'\r'})                // return cursor to start
			s.tw.Stage([]byte("$ " + line))
			s.tw.Commit()

		default:
			fmt.Println("default")
		}
	}
}

func (s *Shell) LookupCommand(name string) (f interpreter.CmdFunc, found bool, err error) {
	cmd, found := s.CommandRegistry.LookupCommand(name)
	if !found {
		return nil, false, nil
	}

	return func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run(cmd, args)
	}, true, nil
}
