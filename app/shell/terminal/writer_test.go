package terminal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceNurse(t *testing.T) {
	input := []byte("echo mino\n mino\n")
	out := bytes.NewBuffer(nil)
	tw := NewTermWriter(out)
	n, err := tw.Write(input)
	assert.NoError(t, err)
	assert.Equal(t, len(input), n)
	assert.Equal(t, "echo mino\r\n mino\r\n", out.String())
}

// func FuzzWrite(f *testing.F) {
// 	f.Add([]byte{keyEscape, csi, keyEnter, keyLF, 'a'})
// 	f.Fuzz(func(t *testing.T, x []byte) {
// 		tw := NewTermWriter(io.Discard)
// 		n, err := tw.Write(x)
// 		if err != nil || n != len(x) {
// 			t.Logf("%q", string(x))
// 			t.Failed()
// 		}
// 	})
// }
