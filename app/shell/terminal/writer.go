package terminal

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

type TermWriter struct {
	w   io.Writer
	buf []byte

	fgColorStack *ColorStack
}

func NewTermWriter(w io.Writer) *TermWriter {
	t := &TermWriter{
		w:            w,
		fgColorStack: newColorStack(),
	}
	t.StagePushForegroundColor(Purple)
	return t
}

func (t *TermWriter) Write(p []byte) (n int, err error) {
	temp := len(t.buf)
	t.Stage(p)
	n, err = t.Commit()
	if err != nil {
		return max(0, n-temp), err
	}
	return len(p), err
}

func (t *TermWriter) WriteString(s string) (n int, err error) {
	return t.Write([]byte(s))
}

func (t *TermWriter) WriteRune(r rune) error {
	b := make([]byte, 8)
	n := utf8.EncodeRune(b, r)
	b = b[:n]
	_, err := t.w.Write(b)
	return err
}

func (t *TermWriter) WriteByte(b byte) error {
	if w, ok := t.w.(io.ByteWriter); ok {
		return w.WriteByte(b)
	}
	_, err := t.w.Write([]byte{b})
	return err
}

func (t *TermWriter) Stagef(format string, a ...any) *TermWriter {
	t.Stage(fmt.Appendf([]byte{}, format, a...))
	return t
}

func (t *TermWriter) Stage(p []byte) *TermWriter {
	// need to replace all \n with \r\n
	// \r is carriage return which places the
	// cursor back all the way to the left
	for iter := 0; len(p) > 0 && iter < 4096; iter++ {
		idx := bytes.IndexRune(p, '\n')
		if idx < 0 {
			t.buf = append(t.buf, p...)
			return t
		}

		t.buf = append(t.buf, p[:idx]...)
		p = p[idx+1:]

		t.buf = append(t.buf, newLine...)
	}

	return t
}

func (t *TermWriter) StageString(s string) *TermWriter {
	t.Stage([]byte(s))
	return t
}

func (t *TermWriter) StageRune(r rune) *TermWriter {
	b := make([]byte, 8)
	n := utf8.EncodeRune(b, r)
	b = b[:n]
	return t.Stage(b)
}

func (t *TermWriter) StageByte(b byte) *TermWriter {
	return t.Stage([]byte{b})
}

func (t *TermWriter) Commit() (int, error) {
	n, err := t.w.Write(t.buf)
	if err != nil {
		t.buf = t.buf[n:]
		return n, err
	}

	t.buf = t.buf[:0]
	return n, nil
}

func (t *TermWriter) StageMove(deltaX int) *TermWriter {
	if deltaX == 0 {
		return t
	}

	var direction byte
	var distance int

	if deltaX > 0 {
		direction = 'C' // right
		distance = deltaX
	} else {
		direction = 'D' // left
		distance = -deltaX
	}

	moveSeq := []byte{keyEscape, '['}
	moveSeq = strconv.AppendInt(moveSeq, int64(distance), 10)
	moveSeq = append(moveSeq, direction)
	return t.Stage(moveSeq)
}

func (t *TermWriter) StagePushForegroundColor(c []byte) *TermWriter {
	t.fgColorStack.Push(c)
	return t.Stage(c)
}

func (t *TermWriter) StagePopForegroundColor() *TermWriter {
	t.fgColorStack.Pop()
	c := t.fgColorStack.Top()
	return t.Stage(c)
}
