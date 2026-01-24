package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

var _ shell.ShellPlugin = (*ClearScreenPlugin)(nil)

type ClearScreenPlugin struct {
	tr *terminal.Terminal
}

func (*ClearScreenPlugin) Name() string {
	return "Clear Screen"
}

func (c *ClearScreenPlugin) Register(s *shell.Shell) {
	c.tr = s.Terminal()
	s.KeyHandlers().Use(terminal.ItemKeyCtrlL, c.onItemKeyCtrlL)
}

func (c *ClearScreenPlugin) onItemKeyCtrlL(next shell.KeyHandler) shell.KeyHandler {
	return func(i terminal.Item) error {
		c.tr.ClearScreen()
		return next(i)
	}
}
