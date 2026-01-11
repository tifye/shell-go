package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipe(t *testing.T) {
	input := `echo 1 | echo 2 | echo 3`
	prog, err := Parse(input)
	require.NoError(t, err)

	Inspect(prog, func(n Node) bool {
		switch n := n.(type) {
		case *PipeStmt:
			assert.Len(t, n.Cmds, 3)
		}
		return true
	})
}

func TestSequential(t *testing.T) {
	input := `echo 1; echo 2 | echo 3;`
	prog, err := Parse(input)
	require.NoError(t, err)

	Inspect(prog, func(n Node) bool {
		switch n := n.(type) {
		case *Root:
			assert.Len(t, n.Cmds, 2)
		}
		return false
	})
}

func TestBackground(t *testing.T) {
	input := `echo 1 | echo 2 &`
	prog, err := Parse(input)
	require.NoError(t, err)

	called := false
	Inspect(prog, func(n Node) bool {
		switch n := n.(type) {
		case *BackgroundStmt:
			called = true
			assert.IsType(t, &PipeStmt{}, n.Stmt)

			return false
		}
		return true
	})

	assert.True(t, called)
}
