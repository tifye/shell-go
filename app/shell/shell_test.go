package shell

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseInput(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		output []string
	}{
		{
			name:   "Spaces are preserved within quotes",
			input:  "'hello     world'",
			output: []string{"hello     world"},
		},
		{
			name:   "Consecutive spaces are collapsed unless quoted",
			input:  "hello     world",
			output: []string{"hello world"},
		},
		{
			name:   "Adjacent quoted strings are concatenated",
			input:  "'hello''world'",
			output: []string{"helloworld"},
		},
		{
			name:   "Empty quotes are ignored",
			input:  "hello''world",
			output: []string{"helloworld"},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(test.input + "\n"))
			args, err := parseInput(reader)
			if err != nil {
				t.Fatalf("expected no err but got: %s", err)
			}

			if len(args) != len(test.output) {
				t.Fatalf("expected output to be %q but got %q", test.output, args)
			}

			for i := range len(test.output) {
				if test.output[i] != args[i] {
					t.Fatalf("expected output to be %q but got %q", test.output, args)
				}
			}
		})
	}
}
