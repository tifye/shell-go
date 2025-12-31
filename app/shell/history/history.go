package history

import (
	"github.com/codecrafters-io/shell-starter-go/assert"
)

type InMemoryHistory struct {
	pos     uint
	history []string
	hooks   []func(string)
}

func NewInMemoryHistory() *InMemoryHistory {
	return &InMemoryHistory{
		history: make([]string, 0),
	}
}

func (h *InMemoryHistory) WithHook(hook func(string)) {
	if h.hooks == nil {
		h.hooks = make([]func(string), 0)
	}
	h.hooks = append(h.hooks, hook)
}

func (h *InMemoryHistory) Add(item string) {
	if len(h.history) > 0 {
		if h.history[len(h.history)-1] == item {
			return
		}
	}
	h.history = append(h.history, item)
	for _, hook := range h.hooks {
		hook(item)
	}
}

func (h *InMemoryHistory) Len() int {
	return len(h.history)
}

func (h *InMemoryHistory) At(idx int) string {
	assert.Assert(idx >= 0)
	assert.Assert(idx < len(h.history))
	return h.history[len(h.history)-1-idx]
}
