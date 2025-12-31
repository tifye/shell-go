package history

import (
	"golang.org/x/term"
)

type HistoryContext struct {
	term.History

	idx int
}

func NewHistoryContext() *HistoryContext {
	return &HistoryContext{
		History: NewInMemoryHistory(),
		idx:     -1,
	}
}

func (h *HistoryContext) Next() string {
	h.idx += 1

	if h.idx >= h.Len() {
		h.idx = h.Len() - 1
	}
	if h.idx < 0 {
		h.idx = 0
	}

	item := h.At(h.idx)
	return item
}

func (h *HistoryContext) Previous() string {
	h.idx -= 1

	if h.idx >= h.Len() {
		h.idx = h.Len() - 1
	}
	if h.idx < 0 {
		h.idx = 0
	}

	item := h.At(h.idx)
	return item
}

func (h *HistoryContext) Reset() {
	h.idx = -1
}
