package shell

import (
	"fmt"
	"regexp"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

type autocompleter struct {
	registry *cmd.Registry
	bellRung bool

	RingTheBell         func()
	PossibleCompletions func([]string)
}

func (a *autocompleter) Complete(input string) (string, bool) {
	escaped := regexp.QuoteMeta(input)
	reg, _ := regexp.Compile(fmt.Sprintf("^(%s)+.*", escaped))
	matches := a.registry.MatchAll(reg)
	fmt.Println(matches)
	if len(matches) == 1 {
		a.bellRung = false
		return matches[0], true
	}

	if !a.bellRung {
		a.RingTheBell()
		a.bellRung = true
		return "", false
	}

	if len(matches) > 1 {
		a.PossibleCompletions(matches)
	}

	return "", false
}
