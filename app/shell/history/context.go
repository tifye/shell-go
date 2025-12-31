package history

import (
	"golang.org/x/term"
)

type HistoryContext struct {
	term.History

	idx  int
	item string
}

func NewHistoryContext(history term.History) *HistoryContext {
	return &HistoryContext{
		History: history,
		idx:     -1,
	}
}

func (h *HistoryContext) Back() bool {
	return h.move(1)
}

func (h *HistoryContext) Forward() bool {
	return h.move(-1)
}

func (h *HistoryContext) move(i int) bool {
	h.idx += i

	if h.idx >= h.Len() {
		h.idx = h.Len() - 1
		return false
	}
	if h.idx < 0 {
		h.idx = 0
		return false
	}

	return true
}

func (h *HistoryContext) Position() int {
	return h.idx
}

func (h *HistoryContext) Reset() {
	h.idx = -1
}

func (h *HistoryContext) Item() string {
	return h.At(h.idx)
}

func (h *HistoryContext) OnAdd(item string) {
	h.Back()
}
