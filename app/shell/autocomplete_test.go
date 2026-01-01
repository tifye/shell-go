package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLongestCommonPrefix(t *testing.T) {
	input := []string{
		"1234",
		"1234566",
		"123455453",
		"1234452425",
		"12344",
		"1234sfsf",
	}
	out := largestCommonPrefix(input)
	assert.Equal(t, "1234", out)
}
