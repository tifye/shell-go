package history

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type openFileFS interface {
	OpenFile(string, int) (io.ReadWriteCloser, error)
}

func ReadHistoryFromFile(h term.History, fsys openFileFS, filename string) error {
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

func WriteHistoryToFile(h term.History, fsys openFileFS, filename string) error {
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

func AppendHistoryToFile(h *HistoryContext, fsys openFileFS, filename string) error {
	if h.Len() == 0 {
		return nil
	}

	file, err := fsys.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	for {
		item, ok := h.Forward()
		if !ok {
			break
		}
		itemb := []byte(item + "\n")
		if _, err := file.Write(itemb); err != nil {
			return fmt.Errorf("fìle write: %w", err)
		}
	}

	return nil
}
