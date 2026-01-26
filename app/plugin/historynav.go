package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
	"golang.org/x/term"
)

var _ shell.ShellPlugin = (*NavHistory)(nil)

type NavHistory struct {
	historyCtx   *history.HistoryContext
	shellHistory term.History
	tr           *terminal.Terminal
}

func NewNavHistory() *NavHistory {
	return &NavHistory{}
}

func (h *NavHistory) Name() string {
	return "Navigate History"
}

func (h *NavHistory) Register(s *shell.Shell) {
	s.AddHook(shell.HookPreRead, h.onPreRead)
	s.KeyHandlers().Use(terminal.ItemKeyUp, h.handleItemUp)
	s.KeyHandlers().Use(terminal.ItemKeyDown, h.handleItemDown)

	h.shellHistory = s.HistoryContext
	h.tr = s.Terminal()
}

func (h *NavHistory) handleItemUp(next shell.KeyHandler) shell.KeyHandler {
	return func(i terminal.Item) error {
		if item, ok := h.historyCtx.Back(); ok {
			h.tr.ReplaceWith(item)
		}
		return next(i)
	}
}

func (h *NavHistory) handleItemDown(next shell.KeyHandler) shell.KeyHandler {
	return func(i terminal.Item) error {
		if item, ok := h.historyCtx.Forward(); ok {
			h.tr.ReplaceWith(item)
		}
		return next(i)
	}
}

func (h *NavHistory) onPreRead() {
	h.historyCtx = history.NewHistoryContext(h.shellHistory)
}
