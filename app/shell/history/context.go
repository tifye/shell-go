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

func (h *HistoryContext) Back() (item string, more bool) {
	h.idx += 1

	if h.idx >= h.Len() {
		h.idx = h.Len() - 1
		more = false
	}
	if h.idx < 0 {
		h.idx = 0
		more = false
	}

	item = h.At(h.idx)
	return
}

func (h *HistoryContext) Forward() (item string, more bool) {
	h.idx -= 1

	if h.idx >= h.Len() {
		h.idx = h.Len() - 1
		more = false
	}
	if h.idx < 0 {
		h.idx = 0
		more = false
	}

	item = h.At(h.idx)
	return
}

func (h *HistoryContext) Position() int {
	return h.idx
}

func (h *HistoryContext) Reset() {
	h.idx = -1
}

func (h *HistoryContext) OnAdd(item string) {
	h.Back()
}
