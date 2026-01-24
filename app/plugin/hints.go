package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

var _ shell.ShellPlugin = (*CompletionHintsPlugin)(nil)

type CompletionHintsPlugin struct {
	tr       *terminal.Terminal
	registry *cmd.Registry
}

func (a *CompletionHintsPlugin) Name() string {
	return "Completion Hints"
}

func (a *CompletionHintsPlugin) Register(s *shell.Shell) {
	tr := s.Terminal()
	a.tr = tr
	a.registry = s.CommandRegistry
	tr.CharacterReadHook = a.onCharacterRead
}

func (a *CompletionHintsPlugin) onCharacterRead(_ rune) {
	line := a.tr.Line()
	if line == "" {
		return
	}

	match, ok := a.registry.MatchFirst(line)
	if !ok {
		return
	}

	suggestions := match[len(line):]

	tw := a.tr.Writer()
	_, _ = tw.Stage(terminal.ClearLine).
		StagePushForegroundColor(terminal.Grey).
		StageString(suggestions).
		StageMove(-len(suggestions)).
		StagePopForegroundColor().
		Commit()
}
