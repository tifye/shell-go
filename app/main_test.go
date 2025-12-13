package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func TestShell(t *testing.T) {
	input := strings.NewReader("type mino\n")
	output := bytes.NewBuffer(nil)

	s := shell.NewShell(output, input, goenv{}, gofs{})
	s.AddBuiltin(builtin.NewTypeCommand(s))
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
}
