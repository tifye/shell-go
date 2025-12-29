package history

import (
	"github.com/codecrafters-io/shell-starter-go/assert"
)

type InMemoryHistory struct {
	pos     uint
	history []string
}

func NewInMemoryHistory() *InMemoryHistory {
	return &InMemoryHistory{
		history: make([]string, 0),
	}
}

func (h *InMemoryHistory) Push(item string) error {
	h.history = append(h.history, item)
	return nil
}

func (h *InMemoryHistory) Next() (string, error) {
	if len(h.history) == 0 {
		return "", ErrHistoryEmpty
	}

	assert.Assert(int(h.pos) < len(h.history))

	nextPos := h.pos + 1
	if int(nextPos) > len(h.history)-1 {
		nextPos = uint(len(h.history) - 1)
	}

	h.pos = nextPos
	return h.history[h.pos], nil
}

func (h *InMemoryHistory) Previous() (string, error) {
	if len(h.history) == 0 {
		return "", ErrHistoryEmpty
	}

	assert.Assert(int(h.pos) < len(h.history))

	nextPos := max(int(h.pos-1), 0)
	h.pos = uint(nextPos)
	return h.history[h.pos], nil
}

func (h *InMemoryHistory) Dump() []string {
	return h.history
}
