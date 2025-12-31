package history_test

import (
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/stretchr/testify/assert"
)

func TestHistoryContext(t *testing.T) {
	h := history.NewInMemoryHistory()
	hctx := history.NewHistoryContext(h)

	h.Add("1")
	h.Add("2")
	h.Add("3")

	item, more := hctx.Forward()
	assert.Equal(t, item, "1")
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, item, "2")
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, item, "3")
	assert.False(t, more)
	//
	item, more = hctx.Forward()
	assert.Equal(t, item, "3")
	assert.False(t, more)

	h.Add("4")
	h.Add("5")
	h.Add("6")

	item, more = hctx.Forward()
	assert.Equal(t, item, "4")
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, item, "5")
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, item, "6")
	assert.False(t, more)
	//
	item, more = hctx.Forward()
	assert.Equal(t, item, "6")
	assert.False(t, more)
}
