package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

type ControlCExitPlugin struct{}

func (ControlCExitPlugin) Name() string {
	return "Control+C Exit"
}

func (ControlCExitPlugin) Register(s *shell.Shell) {
	s.KeyHandlers().Use(terminal.ItemKeyCtrlC, onItemKeyCtrlC)
}

func onItemKeyCtrlC(_ shell.KeyHandler) shell.KeyHandler {
	return func(_ terminal.Item) error {
		return shell.ErrExit
	}
}
