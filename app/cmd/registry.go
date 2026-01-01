package cmd

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

type Registry struct {
	builtins     commandMap
	path         map[string]string
	buildCmdFunc func(exec string, path string) CommandFunc
}

func NewResitry(
	buildCmdFunc func(exec string, path string) CommandFunc,
) *Registry {
	return &Registry{
		builtins:     commandMap{},
		path:         map[string]string{},
		buildCmdFunc: buildCmdFunc,
	}
}

func (r *Registry) AddBuiltinCommand(name string, cmd CommandFunc) {
	assert.NotNil(r.builtins)
	r.builtins[name] = cmd
}

func (r *Registry) AddPathExec(name string, execPath string) {
	assert.NotNil(r.path)
	r.path[name] = execPath
}

func (r *Registry) LookupBuiltinCommand(name string) (*Command, bool) {
	cf, ok := r.builtins.Lookup(name)
	if !ok {
		return nil, false
	}
	return cf(), true
}

func (r *Registry) LookupPathCommand(name string) (string, *Command, bool) {
	execPath, ok := r.path[name]
	if !ok {
		return "", nil, false
	}
	cmd := r.buildCmdFunc(name, execPath)()
	return execPath, cmd, true
}

func (r *Registry) LookupCommand(name string) (*Command, bool) {
	cd, ok := r.builtins.Lookup(name)
	if ok {
		return cd(), true
	}

	_, cmd, ok := r.LookupPathCommand(name)
	if ok {
		return cmd, true
	}

	return nil, false
}

func (r *Registry) MatchFirst(prefix string) (string, bool) {
	for k := range r.builtins {
		if strings.HasPrefix(k, prefix) {
			return k, true
		}
	}

	for k := range r.path {
		if strings.HasPrefix(k, prefix) {
			return k, true
		}
	}

	return "", false
}

type commandMap map[string]CommandFunc

func (m commandMap) Lookup(name string) (CommandFunc, bool) {
	c, ok := m[name]
	return c, ok
}

func LoadFromPathEnv(
	pathEnv string,
	fsys fs.FS,
	fullPath func(string) (string, error),
	buildCmdFunc func(exec string, path string) CommandFunc,
) (*Registry, error) {
	assert.Assert(len(pathEnv) > 0, "expected non-empty pathEnv")

	registry := NewResitry(buildCmdFunc)

	paths := filepath.SplitList(pathEnv)
	for _, p := range paths {
		if len(p) == 0 {
			continue
		}

		cmdPaths, _ := commandsInPath(p, fsys, fullPath)
		for k, v := range cmdPaths {
			registry.AddPathExec(k, v)
		}
	}

	return registry, nil
}

func commandsInPath(dir string, fsys fs.FS, fullPath func(string) (string, error)) (map[string]string, error) {
	assert.Assert(len(dir) > 0, "expected non-empty path")
	assert.Assertf(len(dir) < 4026, "unexpectedly large path: %s", dir)
	assert.NotNil(fsys, "fsys")
	assert.NotNil(fullPath, "fullPath func")

	cmdPaths := map[string]string{}

	dir, _ = fullPath(dir)
	err := fs.WalkDir(fsys, dir,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if path == dir {
				return nil
			}

			if entry.IsDir() {
				return fs.SkipDir
			}

			fi, _ := entry.Info()
			if !hasExecPerms(fi.Mode().Perm()) {
				return nil
			}

			fname := entry.Name()
			fname = strings.TrimSuffix(fname, filepath.Ext(fname))
			if _, ok := cmdPaths[fname]; ok {
				return nil
			}

			cmdPaths[fname] = path
			return nil
		})

	return cmdPaths, err
}

func hasExecPerms(mode fs.FileMode) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return mode&0111 != 0
}
