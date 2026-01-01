package shell

import (
	"io"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

type autocompleter struct {
	registry *cmd.Registry
	w        io.Writer

	bellRung bool
}

func (a *autocompleter) Complete(input string) (string, bool) {
	line, ok := a.registry.MatchFirst(input)
	if ok {
		a.bellRung = false
		return line, true
	}

	if !a.bellRung {
		a.w.Write([]byte{0x07})
		a.bellRung = true
		return "", false
	}

	return "", false
}
