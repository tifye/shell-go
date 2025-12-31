package history

import (
	"golang.org/x/term"
)

type HistoryContext struct {
	term.History

	knownLen int
	pos      int
	item     string
}

func NewHistoryContext(history term.History) *HistoryContext {
	return &HistoryContext{
		History:  history,
		pos:      -1,
		knownLen: history.Len(),
	}
}

func (h *HistoryContext) Back() (string, bool) {
	h.grow()

	if h.Len() == 0 {
		return "", false
	}

	ok := h.forwardIdx() <= h.Len()-1
	if !ok {
		return "", false
	}

	item := h.At(h.backIdx())
	h.advance(1)
	return item, ok
}

func (h *HistoryContext) Forward() (string, bool) {
	h.grow()

	if h.Len() == 0 {
		return "", false
	}

	ok := h.forwardIdx() >= 0
	if !ok {
		return "", false
	}

	item := h.At(h.forwardIdx())
	h.advance(-1)
	return item, ok
}

func (h *HistoryContext) backIdx() int {
	return h.pos + 1
}
func (h *HistoryContext) forwardIdx() int {
	return h.pos - 1
}
func (h *HistoryContext) grow() {
	if grew := h.Len() - h.knownLen; grew > 0 {
		if h.pos == -1 {
			h.pos = 0
		}
		h.pos += grew
		h.knownLen = h.Len()
	}
}
func (h *HistoryContext) advance(n int) {
	h.pos += n
	if h.pos >= h.Len()-1 {
		h.pos = h.Len() - 1
	}
	if h.pos <= 0 {
		h.pos = 0
	}
}

func (h *HistoryContext) Position() int {
	return h.pos
}

func (h *HistoryContext) Reset() {
	h.knownLen = h.Len()
	h.pos = -1
}
