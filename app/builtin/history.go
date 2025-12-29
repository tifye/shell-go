package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/assert"
)

func NewHistoryCommand(s *shell.Shell) *cmd.Command {
	assert.NotNil(s)
	return &cmd.Command{
		Name:   "history",
		Stdin:  s.Stdin,
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		Run: func(cmd *cmd.Command, args []string) error {
			n := -1

			if len(args) >= 2 {
				nArg := args[1]
				nParsed, err := strconv.Atoi(nArg)
				if err != nil {
					return fmt.Errorf("expected integer argument")
				}
				n = nParsed
			}

			hist := s.History.Dump(n)
			for i, item := range hist {
				hist[i] = fmt.Sprintf("\t%d %s", i+1, item)
			}
			histFormatted := strings.Join(hist, "\n")
			_, err := fmt.Fprintln(cmd.Stdout, histFormatted)
			return err
		},
	}
}
