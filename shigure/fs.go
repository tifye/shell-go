package main

import (
	"errors"
	"io/fs"
)

type filesystem struct{}

func (_ filesystem) Open(name string) (fs.File, error) {
	return nil, errors.New("not implemented")
}

func (_ filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}
