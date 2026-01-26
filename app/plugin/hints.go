package plugin

import (
	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

var _ shell.ShellPlugin = (*CompletionHints)(nil)

type CompletionHints struct {
	tr       *terminal.Terminal
	registry *cmd.Registry
}

func NewCompletionHints() *CompletionHints {
	return &CompletionHints{}
}

func (a *CompletionHints) Name() string {
	return "Completion Hints"
}

func (a *CompletionHints) Register(s *shell.Shell) {
	tr := s.Terminal()
	a.tr = tr
	a.registry = s.CommandRegistry
	tr.CharacterReadHook = a.onCharacterRead
}

func (a *CompletionHints) onCharacterRead(_ rune) {
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
