package shell

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
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

type History interface {
	Push(string) error
	Next() (string, error)
	Previous() (string, error)
	Dump(n int) []string
	Size() int64
}

type FS interface {
	fs.ReadDirFS
	OpenFile(string, int) (io.ReadWriteCloser, error)
}

type Shell struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Stdin    io.Reader
	builtins []*cmd.Command
	Env      env
	FS       FS
	Exec     func(cmd *cmd.Command, path string, args []string) error
	FullPath func(string) (string, error)

	HistoryCtx *history.HistoryContext

	tr *terminal.TermReader
	tw *terminal.TermWriter
}

func (s *Shell) Run() error {
	assert.NotNil(s.Stdout)
	assert.NotNil(s.Stderr)
	assert.NotNil(s.Stdin)

	s.tw = terminal.NewTermWriter(s.Stdout)
	s.tr = terminal.NewTermReader(s.Stdin, s.tw)
	s.Stdout = s.tw
	s.Stderr = s.tw

	if histFile := s.Env.Get("HISTFILE"); len(histFile) > 0 {
		err := history.ReadHistoryFromFile(s.HistoryCtx, s.FS, s.Env.Get("HISTFILE"))
		if err != nil {
			fmt.Fprintf(s.Stderr, "failed to load command history: %s\n", err)
		}
	}
	defer func() {
		histFile := s.Env.Get("HISTFILE")
		if len(histFile) <= 0 {
			return
		}
		err := history.AppendHistoryToFile(s.HistoryCtx, s.FS, s.Env.Get("HISTFILE"))
		if err != nil {
			fmt.Fprintf(s.Stderr, "failed to save command history: %s\n", err)
		}
	}()

	for {
		fmt.Fprint(s.Stdout, "$ ")

		input, err := s.read()
		if err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			_, _ = fmt.Fprintf(s.Stdout, "error reading input: %s\n", err)
			return nil
		}
		input = strings.TrimPrefix(input, "$ ")

		s.HistoryCtx.Add(input)

		prog, err := interpreter.Parse(input, s)
		if err != nil {
			if errors.Is(err, interpreter.ErrCommandNotFound) {
				_, _ = fmt.Fprintln(s.Stdout, err)
			} else {
				_, _ = fmt.Fprintf(s.Stdout, "error parsing input: %s\n", err)
			}
			continue
		}

		if err := prog.Run(); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			_, _ = fmt.Fprintf(s.Stderr, "error executing: %s\n", err)
		}
	}
}

func (s *Shell) read() (string, error) {
	// Not exactly optimal but works for now
	histNavCtx := history.NewHistoryContext(s.HistoryCtx.History)

	for {
		switch item := s.tr.NextItem(); item.Type {
		case terminal.ItemKeyUp:
			item, _ := histNavCtx.Back()
			s.tr.ReplaceWith("$ " + item)
		case terminal.ItemKeyDown:
			item, _ := histNavCtx.Forward()
			s.tr.ReplaceWith("$ " + item)
		case terminal.ItemKeyCtrlC:
			return "", ErrExit
		case terminal.ItemLineInput:
			return item.Literal, nil
		default:
			fmt.Println("default")
		}
	}
}

func (s *Shell) AddBuiltins(commands ...*cmd.Command) {
	assert.NotNil(commands)

	if s.builtins == nil {
		s.builtins = make([]*cmd.Command, 0, len(commands))
	}

	for _, cmd := range commands {
		// Ok, small loop
		_, found := s.LookupBuiltinCommand(cmd.Name)
		if !found {
			s.builtins = append(s.builtins, cmd)
		}
	}
}

func (s *Shell) LookupBuiltinCommand(name string) (*cmd.Command, bool) {
	assert.Assert(len(name) > 0)

	for _, c := range s.builtins {
		if strings.EqualFold(c.Name, name) {
			c.Stdin = s.Stdin
			c.Stdout = s.Stdout
			c.Stderr = s.Stderr
			return c, true
		}
	}
	return nil, false
}

func (s *Shell) LookupPathCommand(name string) (string, *cmd.Command, bool) {
	assert.Assert(len(name) > 0)
	assert.NotNil(s.Env)
	assert.NotNil(s.Exec)

	path := s.Env.Get("PATH")
	if len(path) == 0 {
		return "", nil, false
	}

	paths := filepath.SplitList(path)
	for _, p := range paths {
		// todo: what to do with error?
		if len(p) == 0 {
			continue
		}
		exePath, found, err := s.lookupExecutableInDir(p, name)
		if err != nil {
			return "", nil, false
		}
		if found {
			cmd := &cmd.Command{
				Name:   name,
				Stdout: s.Stdout,
				Stderr: s.Stderr,
				Stdin:  s.Stdin,
				Run: func(cmd *cmd.Command, args []string) error {
					return s.Exec(cmd, exePath, args)
				},
			}
			return exePath, cmd, true
		}
	}

	return "", nil, false
}

func (s *Shell) lookupExecutableInDir(dir string, exeName string) (exePath string, found bool, err error) {
	assert.Assert(len(dir) > 0)
	assert.Assert(len(exeName) > 0)

	dir, _ = s.FullPath(dir)

	// todo: not optimal when reading large dirs
	// todo: what to do with error?
	err = fs.WalkDir(s.FS, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if path == dir {
			return nil
		}

		if d.IsDir() {
			return fs.SkipDir
		}

		fname := d.Name()
		fname = strings.TrimSuffix(fname, filepath.Ext(fname))
		if fname != exeName {
			return nil
		}

		fi, _ := d.Info()
		if hasExecPerms(fi.Mode().Perm()) {
			exePath = path
			found = true
			return fs.SkipAll
		}

		return nil
	})

	return exePath, found, err
}

func (s *Shell) LookupCommand(name string) (*cmd.Command, bool, error) {
	assert.Assert(len(name) > 0)

	cmd, found := s.LookupBuiltinCommand(name)
	if found {
		assert.NotNil(cmd)
		return cmd, true, nil
	}

	_, cmd, found = s.LookupPathCommand(name)
	if found {
		return cmd, true, nil
	}

	return nil, false, nil
}
