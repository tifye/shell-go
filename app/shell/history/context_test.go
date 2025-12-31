package history_test

import (
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/stretchr/testify/assert"
)

func TestHistoryContextBack(t *testing.T) {
	h := history.NewInMemoryHistory()
	hctx := history.NewHistoryContext(h)

	h.Add("1")
	h.Add("2")
	h.Add("3")

	hctx.Reset()

	item, more := hctx.Back()
	assert.Equal(t, "3", item)
	assert.True(t, more)
	item, more = hctx.Back()
	assert.Equal(t, "2", item)
	assert.True(t, more)
	item, more = hctx.Back()
	assert.Equal(t, "1", item)
	assert.False(t, more)

	item, more = hctx.Forward()
	assert.Equal(t, "2", item)
	assert.True(t, more)

	item, more = hctx.Back()
	assert.Equal(t, "1", item)
	assert.False(t, more)
}

func TestHistoryContext(t *testing.T) {
	h := history.NewInMemoryHistory()
	hctx := history.NewHistoryContext(h)

	h.Add("1")
	h.Add("2")
	h.Add("3")

	item, more := hctx.Forward()
	assert.Equal(t, "1", item)
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, "2", item)
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, "3", item)
	assert.False(t, more)
	//
	item, more = hctx.Forward()
	assert.Equal(t, "3", item)
	assert.False(t, more)

	h.Add("4")
	h.Add("5")
	h.Add("6")

	item, more = hctx.Forward()
	assert.Equal(t, "4", item)
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, "5", item)
	assert.True(t, more)
	item, more = hctx.Forward()
	assert.Equal(t, "6", item)
	assert.False(t, more)
	//
	item, more = hctx.Forward()
	assert.Equal(t, "6", item)
	assert.False(t, more)
}
