package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

func main() {

	err := shell(os.Stdout, os.Stdin)
	if err != nil {
		panic(err)
	}
}

func shell(w io.Writer, r io.Reader) error {
	assert.NotNil(w)
	assert.NotNil(r)

	reader := bufio.NewReader(r)

	for {
		fmt.Fprint(w, "$ ")
		input, err := reader.ReadBytes('\n')
		if err != nil {
			_, _ = fmt.Fprintf(w, "error reading input: %s\n", err)
			return nil
		}
		input = bytes.TrimRight(input, "\r\n")

		_, _ = fmt.Fprintf(w, "%s: command not found\n", input)
	}

	// infinite program fine for now
	return nil
}
