package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
	"golang.org/x/term"
)

type NavHistoryPlugin struct {
	nextHandler shell.KeyHandler

	historyCtx   *history.HistoryContext
	shellHistory term.History
	tr           *terminal.Terminal
}

func (h *NavHistoryPlugin) Name() string {
	return "Navigate History"
}

func (h *NavHistoryPlugin) Register(s *shell.Shell) {
	s.AddPreReadHook(h.onPreRead)
	s.KeyHandlers().Use(terminal.ItemKeyUp, h)
	s.KeyHandlers().Use(terminal.ItemKeyDown, h)

	h.shellHistory = s.HistoryContext
	h.tr = s.Terminal()
}

func (h *NavHistoryPlugin) Handle(item terminal.Item) {
	switch item.Type {
	case terminal.ItemKeyUp:
		if item, ok := h.historyCtx.Back(); ok {
			h.tr.ReplaceWith(item)
		}
	case terminal.ItemKeyDown:
		if item, ok := h.historyCtx.Forward(); ok {
			h.tr.ReplaceWith(item)
		}
	}

	if h.nextHandler != nil {
		h.nextHandler.Handle(item)
	}
}

func (h *NavHistoryPlugin) Next(k shell.KeyHandler) {
	h.nextHandler = k
}

func (h *NavHistoryPlugin) onPreRead() {
	h.historyCtx = history.NewHistoryContext(h.shellHistory)
}
