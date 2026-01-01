package terminal

import (
	"bytes"
	"io"
	"unicode/utf8"
)

type TermWriter struct {
	w io.Writer
}

func NewTermWriter(w io.Writer) *TermWriter {
	return &TermWriter{
		w: w,
	}
}

func (t *TermWriter) Write(p []byte) (n int, err error) {
	// need to replace all \n with \r\n
	// \r is carriage return which places the
	// cursor back all the way to the left
	for iter := 0; len(p) > 0 && iter < 4096; iter++ {
		idx := bytes.IndexRune(p, '\n')
		if idx < 0 {
			nn, err := t.w.Write(p)
			n += nn
			return n, err
		}

		nn, err := t.w.Write(p[:idx])
		n += nn
		if err != nil {
			return n, err
		}
		p = p[idx+1:]

		_, err = t.w.Write(crlf)
		// n is for how many bytes of
		// the original p that was written
		n += 1
		if err != nil {
			return n, err
		}
	}
	return n, err
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

func (t *TermWriter) Stage(p []byte) {
	t.Write(p)
}
func (t *TermWriter) StageRune(r rune) {
	t.WriteRune(r)
}
func (t *TermWriter) Commit() {

}
