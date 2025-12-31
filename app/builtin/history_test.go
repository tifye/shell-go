package builtin

import (
	"bytes"
	"io"
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/shell/history"
	"github.com/stretchr/testify/assert"
)

func TestAppendHistory(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	hctx := history.NewHistoryContext(history.NewInMemoryHistory())

	hctx.Add("1")
	hctx.Add("2")
	hctx.Add("3")

	err := appendHistoryToFile(hctx, OpenFileFSFunc(func(s string, i int) (io.ReadWriteCloser, error) {
		return &noOpCloser{buf}, nil
	}), "")
	assert.NoError(t, err)

	hctx.Add("4")
	hctx.Add("5")
	hctx.Add("6")

	err = appendHistoryToFile(hctx, OpenFileFSFunc(func(s string, i int) (io.ReadWriteCloser, error) {
		return &noOpCloser{buf}, nil
	}), "")
	assert.NoError(t, err)

	expected := "1\n2\n3\n4\n5\n6\n"
	actual := buf.String()
	assert.Equal(t, expected, actual)
}

type OpenFileFSFunc func(string, int) (io.ReadWriteCloser, error)

func (f OpenFileFSFunc) OpenFile(filename string, flags int) (io.ReadWriteCloser, error) {
	return f(filename, flags)
}

type noOpCloser struct {
	io.ReadWriter
}

func (n *noOpCloser) Close() error {
	return nil
}
