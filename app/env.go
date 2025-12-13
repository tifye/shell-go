package main

import "os"

type goenv struct{}

func (_ goenv) Get(key string) string {
	return os.Getenv(key)
}
