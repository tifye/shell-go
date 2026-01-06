package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
)

type filesystem struct{}

func (_ filesystem) Open(name string) (fs.File, error) {
	return nil, errors.New("not implemented")
}

func (_ filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

func (filesystem) OpenFile(name string, flags int) (io.ReadWriteCloser, error) {
	return nil, os.ErrNotExist
}
