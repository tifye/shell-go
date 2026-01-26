package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

var _ shell.ShellPlugin = (*ClearScreen)(nil)

type ClearScreen struct {
	tr *terminal.Terminal
}

func NewClearScreen() *ClearScreen {
	return &ClearScreen{}
}

func (*ClearScreen) Name() string {
	return "Clear Screen"
}

func (c *ClearScreen) Register(s *shell.Shell) {
	c.tr = s.Terminal()
	s.KeyHandlers().Use(terminal.ItemKeyCtrlL, c.onItemKeyCtrlL)
}

func (c *ClearScreen) onItemKeyCtrlL(next shell.KeyHandler) shell.KeyHandler {
	return func(i terminal.Item) error {
		c.tr.ClearScreen()
		return next(i)
	}
}
