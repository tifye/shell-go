package history

import (
	"golang.org/x/term"
)

type HistoryContext struct {
	term.History

	idx int
}

func NewHistoryContext(history term.History) *HistoryContext {
	return &HistoryContext{
		History: history,
		idx:     -1,
	}
}

func (h *HistoryContext) Back() string {
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

func (h *HistoryContext) Forward() string {
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

func (h *HistoryContext) Position() int {
	return h.idx
}

func (h *HistoryContext) Reset() {
	h.idx = -1
}

func (h *HistoryContext) Add(item string) {
	h.idx += 1
	h.History.Add(item)
}
