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
			assert.Len(t, n.Chain, 3)
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
		case *Program:
			assert.Len(t, n.Chain, 2)
		}
		return false
	})
}
