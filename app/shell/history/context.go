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
		pos:      0,
		knownLen: history.Len(),
	}
}

func (h *HistoryContext) Back() (string, bool) {
	if grew := h.Len() - h.knownLen; grew > 0 {
		h.pos += grew - 1
		h.knownLen = h.Len()
	}

	item := h.At(h.pos)

	h.pos += 1
	if h.pos > h.Len()-1 {
		h.pos = h.Len() - 1
		return item, false
	}

	return item, true
}

func (h *HistoryContext) Forward() (string, bool) {
	if grew := h.Len() - h.knownLen; grew > 0 {
		h.pos += grew - 1
		h.knownLen = h.Len()
	}

	item := h.At(h.pos)

	h.pos -= 1
	if h.pos < 0 {
		h.pos = 0
		return item, false
	}

	return item, true
}

func (h *HistoryContext) Position() int {
	return h.pos
}

func (h *HistoryContext) Reset() {
	h.knownLen = h.Len()
	h.pos = 0
}
