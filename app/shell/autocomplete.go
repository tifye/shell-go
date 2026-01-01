package shell

import "github.com/codecrafters-io/shell-starter-go/app/cmd"

type autocompleter struct {
	registry *cmd.Registry
}

func (a *autocompleter) Complete(input string) (string, bool) {
	return a.registry.MatchFirst(input)
}
