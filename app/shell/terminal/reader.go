package terminal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

// ANSI escape sequences
// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
const (
	keyCtrlC = 3 // ^C
	keyCtrlD = 4
	keyCtrlU = 21
	keyEnter = '\r'
	keyLF    = '\n'

	keyEscape byte = 0x1B // 27
	// Control Sequence Introducer
	csi byte = 0x5B // '['

	keyBackspace = 127

	keyUnknown = 0xd800 /* UTF-16 surrogate area */ + iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyAltLeft
	keyAltRight
	keyHome
	keyEnd
	keyDeleteWord
	keyDeleteLine
	keyClearScreen
	keyPasteStart
	keyPasteEnd
)

var (
	crlf       = []byte{'\r', '\n'}
	pasteStart = []byte{keyEscape, '[', '2', '0', '0', '~'}
	pasteEnd   = []byte{keyEscape, '[', '2', '0', '1', '~'}
)

type ItemType int

const (
	ItemError ItemType = iota
	ItemEOF
	ItemLineInput
	ItemKeyUp
	ItemKeyDown
	ItemKeyCtrlC
	ItemKeyUnknown
)

type Item struct {
	Type    ItemType
	Literal string
}

type stateFunc func(*TermReader) stateFunc

type TermReader struct {
	r  io.Reader
	tw *TermWriter

	line []rune
	item Item
	view []byte
	buf  [256]byte
}

func NewTermReader(r io.Reader, tw *TermWriter) *TermReader {
	return &TermReader{
		r:  r,
		tw: tw,
	}
}

func (t *TermReader) NextItem() Item {
	t.item = Item{
		Type:    ItemEOF,
		Literal: "EOF",
	}

	state := advance
	for {
		state = state(t)
		if state == nil {
			return t.item
		}
	}
}

func (t *TermReader) ReplaceWith(input string) error {
	t.line = []rune(input)
	t.tw.WriteByte('\r')
	t.tw.Write([]byte{keyEscape, csi, 'K'})
	t.tw.WriteString(input)
	return nil
}

func (t *TermReader) error(e error) stateFunc {
	t.item = Item{
		Type:    ItemError,
		Literal: e.Error(),
	}
	return nil
}

func (t *TermReader) emitItem(item Item) stateFunc {
	t.item = item
	return nil
}

func (t *TermReader) emit(typ ItemType, literal string) stateFunc {
	return t.emitItem(Item{
		Type:    typ,
		Literal: literal,
	})
}

func (t *TermReader) advanceView(n int) {
	assert.Assert(n >= 0)
	assert.Assert(n <= len(t.view))
	t.view = t.view[n:]
}

func (t *TermReader) isViewCurrent(r byte) bool {
	if len(t.view) == 0 {
		return false
	}
	return t.view[0] == r
}

func advance(t *TermReader) stateFunc {
	if len(t.view) > 0 {
		// moves view to front of bof
		// effectively removing the already read parts
		// of the view
		n := copy(t.buf[:], t.view)
		t.view = t.buf[:n]
	}

	readBuf := t.buf[len(t.view):]

	n, err := t.r.Read(readBuf)
	if err != nil && !errors.Is(err, io.EOF) {
		return t.error(fmt.Errorf("advance: %w", err))
	}

	t.view = t.buf[:n+len(t.view)]
	return readInput
}

func readInput(t *TermReader) stateFunc {
	return readKey
}

func readKey(t *TermReader) stateFunc {
	if len(t.view) == 0 {
		return advance
	}

	// https://i.sstatic.net/X9e5B.png
	// From ascii control char table
	switch b := t.view[0]; b {
	case keyCtrlC:
		t.advanceView(1)
		return t.emit(ItemKeyCtrlC, string(b))
	case 14: // ^N
		t.advanceView(1)
		return t.emit(ItemKeyDown, string(b))
	case 16: // ^P
		t.advanceView(1)
		return t.emit(ItemKeyUp, string(b))
	}

	if t.view[0] != keyEscape {
		return handleKey
	}

	if len(t.view) >= 3 {
		if t.view[0] == keyEscape && t.view[1] == csi {
			// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797#cursor-controls
			switch b := t.view[2]; b {
			case 'A':
				t.advanceView(3)
				return t.emit(ItemKeyUp, string(b))
			case 'B':
				t.advanceView(3)
				return t.emit(ItemKeyDown, string(b))
			}
		}
	}

	if len(t.view) >= 6 {
		if bytes.Equal(t.view[:6], pasteStart) {
			t.advanceView(6)
			return readPaste
		}
	}

	return handleKey
}

func readPaste(t *TermReader) stateFunc {
	return nil
}

func handleKey(t *TermReader) stateFunc {
	if !utf8.FullRune(t.view) {
		return advance
	}

	key, size := utf8.DecodeRune(t.view)
	t.advanceView(size)

	switch key {
	case keyEnter, keyLF:
		return handleEnterKey
	default:
		if key >= 32 {
			return t.addToLine(key)
		}
		return advance
	}
}

func handleEnterKey(t *TermReader) stateFunc {
	if t.isViewCurrent(keyLF) {
		t.advanceView(1)
	}
	line := string(t.line)
	t.line = t.line[:0]
	t.tw.Stage(crlf)
	t.tw.Commit()
	return t.emit(ItemLineInput, line)
}

func (t *TermReader) addToLine(r rune) stateFunc {
	t.line = append(t.line, r)
	t.tw.StageRune(r)
	t.tw.Commit()
	return readInput
}
