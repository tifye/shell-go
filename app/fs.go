package main

import (
	"io"
	"io/fs"
	"os"
)

type gofs struct{}

func (_ gofs) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (_ gofs) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (_ gofs) OpenFile(name string, flags int) (io.ReadWriteCloser, error) {
	return os.OpenFile(name, flags, 0644)
}
