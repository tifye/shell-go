package builtin

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"golang.org/x/term"
)

type historyOptions struct {
	readFilename string
}

func NewHistoryCommand(historyCtx *history.HistoryContext, fsys fs.FS) *cmd.Command {
	return &cmd.Command{
		Name: "history",
		Run: func(cmd *cmd.Command, args []string) error {

			opts := &historyOptions{}
			flagset := flag.NewFlagSet("history", flag.ExitOnError)
			flagset.StringVar(&opts.readFilename, "r", "", "Path to file from which to read history entries")
			if err := flagset.Parse(args[1:]); err != nil {
				return fmt.Errorf("parsing args: %w", err)
			}
			args = flagset.Args()

			if len(opts.readFilename) > 0 {
				return readHistoryFromFile(historyCtx, fsys, opts.readFilename)
			}

			numItems := historyCtx.Len()
			if nArg := flagset.Arg(0); len(nArg) > 0 {
				nParsed, err := strconv.Atoi(nArg)
				if err != nil {
					return fmt.Errorf("expected integer argument")
				}
				if nParsed < numItems {
					numItems = nParsed
				}
			}

			hist := make([]string, numItems)
			for i := range numItems {
				hist[i] = historyCtx.At(i)
			}
			// put most recent ones last
			slices.Reverse(hist)

			offset := historyCtx.Len() - numItems
			for i, item := range hist {
				hist[i] = fmt.Sprintf("  %d %s", offset+i+1, item)
			}
			histFormatted := strings.Join(hist, "\n")
			_, err := fmt.Fprintln(cmd.Stdout, histFormatted)
			return err
		},
	}
}

func readHistoryFromFile(h term.History, fsys fs.FS, filename string) error {
	file, err := fsys.Open(filename)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		h.Add(line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	return nil
}
