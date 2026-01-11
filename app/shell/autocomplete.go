package shell

import (
	"fmt"
	"regexp"
	"slices"

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

	a.PossibleCompletions(matches)
	return "", false
}

// MatchFirst tried to complete the input and
// returns the first match it finds. It does not
// ring the bell or check any possible completions.
func (a *autocompleter) MatchFirst(input string) (string, bool) {
	escaped := regexp.QuoteMeta(input)
	reg, _ := regexp.Compile(fmt.Sprintf("^(%s)+.*", escaped))
	matches := a.registry.MatchAll(reg)
	if len(matches) > 0 {
		a.bellRung = false
		return matches[0], true
	}

	return "", false
}

func (a *autocompleter) ringTheBell() bool {
	if a.bellRung {
		return false
	}
	a.RingTheBell()
	a.bellRung = true
	return true
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
