package fsys

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type File io.ReadWriteCloser

type Fsys interface {
	fs.ReadDirFS

	OpenFile(string, int) (io.ReadWriteCloser, error)

	FullPath(string) (string, error)
}

type OSFsys struct {
	WorkingDir string
}

func (f *OSFsys) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (f *OSFsys) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (f *OSFsys) OpenFile(name string, flags int) (File, error) {
	return os.OpenFile(name, flags, 0644)
}

func (f *OSFsys) FullPath(p string) (string, error) {
	return filepath.Abs(p)
}
