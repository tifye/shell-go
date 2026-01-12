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
	keyCtrlC          = 3 // ^C
	keyCtrlD          = 4
	keyCtrlU          = 21
	keyBackspace      = 8
	keyDelete         = 127
	keyCarriageReturn = '\r'
	keyLineFeed       = '\n'
	keyTab            = '\t'

	keyEscape byte = 0x1B // 27
	// Control Sequence Introducer
	csi byte = 0x5B // '['
)

var (
	newLine = []byte{'\r', '\n'}

	// clears line starting from cursor position to end of line
	clearLine   = []byte{keyEscape, csi, 'K'}
	clearScreen = []byte{keyEscape, '[', '2', 'J'}

	pasteStart = []byte{keyEscape, '[', '2', '0', '0', '~'}
	pasteEnd   = []byte{keyEscape, '[', '2', '0', '1', '~'}

	resetColor   = []byte{keyEscape, '[', '0', 'm'}
	Purple       = []byte{keyEscape, '[', '3', '8', ';', '5', ';', '1', '4', '1', 'm'}
	Grey         = []byte{keyEscape, '[', '9', '0', 'm'}
	Cyan         = []byte{keyEscape, '[', '3', '6', 'm'}
	PastelRed    = []byte("\x1b[38;2;255;140;140m") // soft pink-red
	Salmon       = []byte("\x1b[38;2;255;160;122m") // salmon / coral-ish
	Rose         = []byte("\x1b[38;2;255;120;170m") // rosy magenta-red
	Red          = []byte{0x1b, '[', '3', '1', 'm'}
	OffWhiteWarm = []byte("\x1b[38;2;245;244;240m") // warm paper
	OffWhiteCool = []byte("\x1b[38;2;236;239;244m") // cool soft gray
	Ivory        = []byte("\x1b[38;2;255;252;240m") // ivory (very light)
)

type ItemType int

const (
	ItemError ItemType = iota
	ItemEOF
	ItemLineInput
	ItemKeyUp
	ItemKeyDown
	ItemKeyCtrlC
	ItemKeyCtrlL
	ItemKeyTab
	ItemBackspace
	ItemKeyUnknown
)

type Item struct {
	Type    ItemType
	Literal string
}

type stateFunc func(*Terminal) stateFunc

type Terminal struct {
	r      io.Reader
	tw     *TermWriter
	prompt string

	// line is the current user input
	line []rune
	item Item
	view []byte
	buf  [256]byte

	CharacterReadHook func(r rune)
}

func NewTermReader(r io.Reader, tw *TermWriter) *Terminal {
	return &Terminal{
		prompt: "$ ",
		r:      r,
		tw:     tw,
	}
}

func (t *Terminal) Line() string {
	// naughty naughty
	// return strings.TrimPrefix(string(t.line), string(t.prompt)+" ")
	return string(t.line)
}

func (t *Terminal) NextItem() Item {
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

func (t *Terminal) ReplaceWith(input string) error {
	t.line = []rune(input)
	t.tw.StageByte(keyCarriageReturn)
	t.tw.StageString(t.prompt)
	t.tw.Stage(clearLine)
	t.tw.StageString(input)
	t.tw.Commit()
	return nil
}

func (t *Terminal) Suggest(s string) error {
	t.tw.Stage(clearLine)
	t.tw.StagePushForegroundColor(Grey)
	t.tw.StageString(s)
	t.tw.StageMove(-len(s))
	t.tw.StagePopForegroundColor()
	_, err := t.tw.Commit()
	return err
}

func (t *Terminal) Ready() error {
	t.tw.StageByte(keyCarriageReturn)
	t.tw.StageString(t.prompt)
	if len(t.line) > 0 {
		t.tw.StageString(t.Line())
	}
	_, err := t.tw.Commit()
	return err
}

func (t *Terminal) ClearScreen() error {
	t.tw.Stage(clearScreen)
	t.tw.StageByte(keyCarriageReturn)
	return t.Ready()
}

func (t *Terminal) eraseLastKey() {
	if len(t.line) == 0 {
		return
	}

	t.line = t.line[:len(t.line)-1]
	t.tw.Stage([]byte{0x1b, '[', '1', 'D'})
	t.tw.Stage([]byte{0x1b, '[', '0', 'K'})
	t.tw.Commit()
}

func (t *Terminal) error(e error) stateFunc {
	t.item = Item{
		Type:    ItemError,
		Literal: e.Error(),
	}
	return nil
}

func (t *Terminal) emitItem(item Item) stateFunc {
	t.item = item
	return nil
}

func (t *Terminal) emit(typ ItemType, literal string) stateFunc {
	return t.emitItem(Item{
		Type:    typ,
		Literal: literal,
	})
}

func (t *Terminal) advanceView(n int) {
	assert.Assert(n >= 0)
	assert.Assert(n <= len(t.view))
	t.view = t.view[n:]
}

func (t *Terminal) isViewCurrent(r byte) bool {
	if len(t.view) == 0 {
		return false
	}
	return t.view[0] == r
}

func advance(t *Terminal) stateFunc {
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

func readInput(t *Terminal) stateFunc {
	return readKey
}

func readKey(t *Terminal) stateFunc {
	if len(t.view) == 0 {
		return advance
	}

	// https://i.sstatic.net/X9e5B.png
	// From ascii control char table
	switch b := t.view[0]; b {
	case keyCtrlC:
		t.advanceView(1)
		return t.emit(ItemKeyCtrlC, string(b))
	case keyBackspace, keyDelete:
		return handleKey
	case 12: // ^L
		t.advanceView(1)
		return t.emit(ItemKeyCtrlL, string(b))
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

func readPaste(t *Terminal) stateFunc {
	return nil
}

func handleKey(t *Terminal) stateFunc {
	if !utf8.FullRune(t.view) {
		return advance
	}

	key, size := utf8.DecodeRune(t.view)
	t.advanceView(size)

	switch key {
	case keyTab:
		return t.emit(ItemKeyTab, string(key))
	case keyCarriageReturn, keyLineFeed:
		return handleEnterKey
	case keyBackspace, keyDelete:
		t.eraseLastKey()
		return readInput
	default:
		if key >= 32 {
			return t.addToLine(key)
		}
		return advance
	}
}

func handleEnterKey(t *Terminal) stateFunc {
	if t.isViewCurrent(keyLineFeed) {
		t.advanceView(1)
	}
	line := string(t.line)
	t.line = t.line[:0]
	t.tw.Stage(newLine)
	t.tw.Commit()
	return t.emit(ItemLineInput, line)
}

func (t *Terminal) addToLine(r rune) stateFunc {
	t.line = append(t.line, r)
	t.tw.StageRune(r)
	t.tw.Commit()

	if t.CharacterReadHook != nil {
		t.CharacterReadHook(r)
	}
	return readInput
}
