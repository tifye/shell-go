package builtin

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

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
				return readHistoryFromFile(historyCtx, fsys, opts.readFilename)
			case len(opts.writeFilename) > 0:
				return writeHistoryToFile(historyCtx, fsys, opts.writeFilename)
			case len(opts.appendFilename) > 0:
				return appendHistoryToFile(historyCtx, fsys, opts.appendFilename)
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

func readHistoryFromFile(h term.History, fsys OpenFileFS, filename string) error {
	file, err := fsys.OpenFile(filename, os.O_RDONLY)
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

func writeHistoryToFile(h term.History, fsys OpenFileFS, filename string) error {
	file, err := fsys.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	for i := range h.Len() {
		if _, err := file.Write([]byte(h.At(h.Len()-1-i) + "\n")); err != nil {
			return fmt.Errorf("file write: %w", err)
		}
	}

	return nil
}

func appendHistoryToFile(h *history.HistoryContext, fsys OpenFileFS, filename string) error {
	file, err := fsys.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	for h.Position() > 0 {
		item := []byte(h.Forward() + "\n")
		if _, err := file.Write(item); err != nil {
			return fmt.Errorf("fìle write: %w", err)
		}
	}

	return nil
}

func printHistory(h term.History, w io.Writer, n int) error {
	hist := make([]string, n)
	for i := range n {
		hist[i] = h.At(i)
	}
	// put most recent ones last
	slices.Reverse(hist)

	offset := h.Len() - n
	for i, item := range hist {
		hist[i] = fmt.Sprintf("  %d %s", offset+i+1, item)
	}
	histFormatted := strings.Join(hist, "\n")
	_, err := fmt.Fprintln(w, histFormatted)
	return err
}
