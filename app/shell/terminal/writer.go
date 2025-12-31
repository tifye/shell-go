package terminal

import "io"

type TermWriter struct {
	w io.Writer
}

func NewTermWriter(w io.Writer) *TermWriter {
	return &TermWriter{
		w: w,
	}
}

func (t *TermWriter) Write(p []byte) (int, error) {
	return t.w.Write(p)
}

func (t *TermWriter) WriteRune(r rune) error {
	_, err := t.w.Write([]byte(string(r)))
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
