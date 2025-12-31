package terminal

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextToken(t *testing.T) {
	tt := []struct {
		name  string
		input []byte
		items []Item
	}{
		{
			"simple with linefeed",
			imp("echo mino", keyEnter),
			[]Item{{ItemLineInput, "echo mino"}},
		},
		{
			"left right keys then simple input",
			imp(keyEscape, "[D", keyEscape, "[C", "echo mino", keyEnter),
			[]Item{{ItemLineInput, "[D[Cecho mino"}},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			tr := NewTermReader(bytes.NewReader(test.input), NewTermWriter(io.Discard))
			for i := range test.items {
				item := tr.NextItem()
				assert.Equal(t, test.items[i].Type, item.Type)
				assert.Equal(t, test.items[i].Literal, item.Literal)
			}
		})

	}
}

func imp(args ...any) []byte {
	input := make([]byte, 0, len(args))
	for i := range args {
		switch a := args[i].(type) {
		case []byte:
			input = append(input, a...)
		case rune:
			input = append(input, []byte(string(a))...)
		case byte:
			input = append(input, a)
		case string:
			input = append(input, []byte(a)...)
		default:
			panic("unsupported type")
		}
	}
	return input
}
