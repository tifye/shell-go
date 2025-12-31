package terminal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceNurse(t *testing.T) {
	input := []byte("echo mino\n")
	out := bytes.NewBuffer(nil)
	tw := NewTermWriter(out)
	n, err := tw.Write(input)
	assert.NoError(t, err)
	assert.Equal(t, len(input)+1, n)
	assert.Equal(t, "echo mino\r\n", out.String())
}
