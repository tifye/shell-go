package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestShell(t *testing.T) {
	input := strings.NewReader("mino\r\n")
	output := bytes.NewBuffer(nil)

	if err := shell(output, input); err != nil {
		t.Fatal(err)
	}

	expected := `$ mino: command not found`

	result := output.String()
	if result != expected {
		t.Fatalf("expected %s but got %s", expected, result)
	}
}
