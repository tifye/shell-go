package plugin

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
)

var _ shell.ShellPlugin = (*AutocompletePlugin)(nil)

type AutocompletePlugin struct {
	tr       *terminal.Terminal
	registry *cmd.Registry

	bellRung bool
}

func (a *AutocompletePlugin) Name() string {
	return "Autocomplete"
}

func (a *AutocompletePlugin) Register(s *shell.Shell) {
	a.registry = s.CommandRegistry
	a.tr = s.Terminal()
	s.KeyHandlers().Use(terminal.ItemKeyTab, a.handleItemKeyTab)
}

func (a *AutocompletePlugin) handleItemKeyTab(next shell.KeyHandler) shell.KeyHandler {
	return func(i terminal.Item) error {
		if line, ok := a.complete(a.tr.Line()); ok {
			a.tr.ReplaceWith(line)
		}
		return next(i)
	}
}

func (a *AutocompletePlugin) complete(input string) (string, bool) {
	escaped := regexp.QuoteMeta(input)
	reg, _ := regexp.Compile(fmt.Sprintf("^(%s)+.*", escaped))
	matches := a.registry.MatchAll(reg)
	if len(matches) == 1 {
		a.bellRung = false
		return matches[0] + " ", true
	}

	if len(matches) == 0 {
		a.ringTheBell()
		return "", false
	}

	prefix := largestCommonPrefix(matches)
	if prefix != input {
		return prefix, true
	} else {
		if a.ringTheBell() {
			return "", false
		}
	}

	a.printPossibleCompletions(matches)
	return "", false
}

func (a *AutocompletePlugin) ringTheBell() bool {
	if a.bellRung {
		return false
	}

	a.tr.Writer().Write([]byte{0x07})
	a.bellRung = true
	return true
}

func (a *AutocompletePlugin) printPossibleCompletions(completions []string) {
	a.tr.Writer().StagePushForegroundColor(terminal.Cyan).
		Stagef("\n%s\n", strings.Join(completions, "  ")).
		StagePopForegroundColor()
	a.tr.Ready()
}

// MatchFirst tries to complete the input and
// returns the first match it finds. It does not
// ring the bell or check any possible completions.
func (a *AutocompletePlugin) MatchFirst(input string) (string, bool) {
	escaped := regexp.QuoteMeta(input)
	reg, _ := regexp.Compile(fmt.Sprintf("^(%s)+.*", escaped))
	matches := a.registry.MatchAll(reg)
	if len(matches) > 0 {
		a.bellRung = false
		return matches[0], true
	}

	return "", false
}

func largestCommonPrefix(s []string) string {
	slices.Sort(s)
	a := s[0]
	b := s[len(s)-1]
	n := min(len(a), len(b))
	buf := make([]byte, 0, n)
	for i := range n {
		if a[i]^b[i] == 0 {
			buf = append(buf, a[i])
		} else {
			break
		}
	}
	return string(buf)
}
