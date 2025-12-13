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
}

func NewShell(w io.Writer, r io.Reader, env env, fs fs.ReadDirFS) *Shell {
	assert.NotNil(w)
	assert.NotNil(r)
	assert.NotNil(env)
	assert.NotNil(fs)
	return &Shell{
		Stdout:   w,
		Stdin:    r,
		Env:      env,
		FS:       fs,
		builtins: make([]*cmd.Command, 0),
	}
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

func (s *Shell) AddBuiltin(command *cmd.Command) {
	assert.NotNil(command)
	s.builtins = append(s.builtins, command)
}

func (s *Shell) LookupBuiltinCommand(name string) (*cmd.Command, bool) {
	assert.Assert(len(name) > 0)

	for _, c := range s.builtins {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

func (s *Shell) LookupPathCommand(name string) (string, *cmd.Command, bool) {
	assert.Assert(len(name) > 0)

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
				Name: name,
				Run: func(args []string) error {
					fmt.Fprintf(s.Stdout, "%s\n", exePath)
					return nil
				},
			}
			return exePath, cmd, true
		}
	}

	return "", nil, false
}

func (s *Shell) lookupExecutableInDir(dir string, exeName string) (exePath string, found bool, err error) {
	assert.Assert(len(dir) > 0)

	if !fs.ValidPath(dir) {
		return
	}

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
		if fi.Mode().Perm()&0111 == 0 {
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
