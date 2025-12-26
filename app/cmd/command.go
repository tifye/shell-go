package cmd

import "io"

type CommandRunFunc func(cmd *Command, args []string) error

type Command struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Name   string
	Run    CommandRunFunc
}
