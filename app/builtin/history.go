package builtin

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewHistoryCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name: "history",
		Run: func(cmd *cmd.Command, args []string) error {
			n := s.HistoryCtx.Len()

			if len(args) >= 2 {
				nArg := args[1]
				nParsed, err := strconv.Atoi(nArg)
				if err != nil {
					return fmt.Errorf("expected integer argument")
				}
				if nParsed < n {
					n = nParsed
				}
				n = nParsed
			}

			hist := make([]string, n)
			for i := range n {
				hist[i] = s.HistoryCtx.At(i)
			}
			// put most recent ones last
			slices.Reverse(hist)

			offset := s.HistoryCtx.Len() - n
			for i, item := range hist {
				hist[i] = fmt.Sprintf("  %d %s", offset+i+1, item)
			}
			histFormatted := strings.Join(hist, "\n")
			_, err := fmt.Fprintln(cmd.Stdout, histFormatted)
			return err
		},
	}
}
