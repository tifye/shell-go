package builtin

import (
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"golang.org/x/term"
)

type OpenFileFS interface {
	OpenFile(string, int) (io.ReadWriteCloser, error)
}

type historyOptions struct {
	readFilename   string
	writeFilename  string
	appendFilename string
}

func NewHistoryCommand(historyCtx *history.HistoryContext, fsys OpenFileFS) *cmd.Command {
	return &cmd.Command{
		Name: "history",
		Run: func(cmd *cmd.Command, args []string) error {

			opts := &historyOptions{}
			flagset := flag.NewFlagSet("history", flag.ExitOnError)
			flagset.StringVar(&opts.readFilename, "r", "", "Path to file from which to read history entries")
			flagset.StringVar(&opts.writeFilename, "w", "", "Path to file to which to write history entries")
			flagset.StringVar(&opts.appendFilename, "a", "", "Path to file to which to append history entries")
			if err := flagset.Parse(args[1:]); err != nil {
				return fmt.Errorf("parsing args: %w", err)
			}
			args = flagset.Args()

			switch {
			case len(opts.readFilename) > 0:
				return history.ReadHistoryFromFile(historyCtx, fsys, opts.readFilename)
			case len(opts.writeFilename) > 0:
				return history.WriteHistoryToFile(historyCtx, fsys, opts.writeFilename)
			case len(opts.appendFilename) > 0:
				return history.AppendHistoryToFile(historyCtx, fsys, opts.appendFilename)
			default:
				numItems := historyCtx.Len()
				if nArg := flagset.Arg(0); flagset.NArg() > 0 {
					nParsed, err := strconv.Atoi(nArg)
					if err != nil {
						return fmt.Errorf("expected integer argument")
					}
					if nParsed < numItems {
						numItems = nParsed
					}
				}
				return printHistory(historyCtx, cmd.Stdout, numItems)
			}
		},
	}
}

func printHistory(h term.History, w io.Writer, n int) error {
	if n <= 0 {
		return nil
	}

	if h.Len() < n {
		n = h.Len()
	}

	offset := h.Len() - n
	for i := range n {
		item := []byte(h.At(i))
		if _, err := fmt.Fprintf(w, "  %d  %s\n", offset+i, item); err != nil {
			return err
		}
	}
	return nil
}
