package shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
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

type Shell struct {
	Stdout   io.Writer
	Stdin    io.Reader
	builtins []*cmd.Command
	Env      env
	FS       fs.ReadDirFS
	FullPath func(string) (string, error)
}

func (s *Shell) Run() error {
	assert.NotNil(s.Stdout)
	assert.NotNil(s.Stdin)

	reader := bufio.NewReader(s.Stdin)

	for {
		fmt.Fprint(s.Stdout, "$ ")
		input, err := reader.ReadBytes('\n')
		if err != nil {
			_, _ = fmt.Fprintf(s.Stdout, "error reading input: %s\n", err)
			return nil
		}

		input = bytes.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue
		}

		args := strings.Fields(string(input))
		cmdName := args[0]
		cmd, found, err := s.LookupCommand(cmdName)
		if err != nil {
			_, _ = fmt.Fprintf(s.Stdout, "error looking up command '%s': %s\n", cmdName, err)
			continue
		}
		if !found {
			_, _ = fmt.Fprintf(s.Stdout, "%s: command not found\n", cmdName)
			continue
		}

		assert.NotNil(cmd)
		if err = cmd.Run(args); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}

			_, _ = fmt.Fprintf(s.Stdout, "error executing '%s': %s\n", input, err)
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
			return c, true
		}
	}
	return nil, false
}

func (s *Shell) LookupPathCommand(name string) (string, *cmd.Command, bool) {
	assert.Assert(len(name) > 0)
	assert.NotNil(s.Env)

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
			cmd := newExecCommand(s, name, exePath)
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
