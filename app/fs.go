package main

import (
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
