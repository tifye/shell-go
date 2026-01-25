package interpreter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpreter(t *testing.T) {
	input := `echo 1 | echo 2; echo 3 ${4}`
	outBuf := bytes.NewBuffer(nil)
	interp := NewInterpreter(
		WithIO(strings.NewReader(input), outBuf, outBuf),
		WithEnvFunc(func(s string) string {
			assert.Equal(t, "4", s)
			return s
		}),
		WithCmdLookupFunc(func(name string) (cmd CmdFunc, found bool, err error) {
			assert.Equal(t, "echo", name)
			return func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
				assert.NotNil(t, stdin)
				assert.NotNil(t, stderr)
				if assert.NotNil(t, stdout) {
					fmt.Fprintln(stdout, strings.Join(args, " "))
				}
				return nil
			}, true, nil
		}),
	)

	synctest.Test(t, func(t *testing.T) {
		err := interp.Evaluate(input)
		require.NoError(t, err)

		output := outBuf.String()
		expected := "echo 2\necho 3 4\n"
		assert.Equal(t, expected, output)
	})
}

func TestPipeline(t *testing.T) {
	input := `echo 1 | echo 2`
	outBuf := bytes.NewBuffer(nil)
	interpStdin := strings.NewReader(input)
	interp := NewInterpreter(
		WithIO(interpStdin, outBuf, outBuf),
		WithCmdLookupFunc(func(name string) (cmd CmdFunc, found bool, err error) {
			assert.Equal(t, "echo", name)
			return func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
				assert.NotNil(t, stdin)
				assert.NotNil(t, stderr)

				if stdin != interpStdin {
					_, err := io.Copy(stdout, stdin)
					assert.NoError(t, err)
				}

				if assert.NotNil(t, stdout) {
					fmt.Fprintln(stdout, strings.Join(args, " "))
				}
				return nil
			}, true, nil
		}),
	)

	synctest.Test(t, func(t *testing.T) {
		err := interp.Evaluate(input)
		require.NoError(t, err)

		output := outBuf.String()
		expected := "echo 1\necho 2\n"
		assert.Equal(t, expected, output)
	})
}

func TestRedirect(t *testing.T) {
	input := `echo 1 >> test; echo 2 > test`
	outBuf := bytes.NewBuffer(nil)
	interp := NewInterpreter(
		WithCmdLookupFunc(func(name string) (cmd CmdFunc, found bool, err error) {
			assert.Equal(t, "echo", name)
			return func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
				assert.NotNil(t, stdin)
				assert.NotNil(t, stderr)
				if assert.NotNil(t, stdout) {
					fmt.Fprintln(stdout, strings.Join(args, " "))
				}
				return nil
			}, true, nil
		}),
		WithOpenFileFunc(func(s string, i int, fm os.FileMode) (io.ReadWriteCloser, error) {
			assert.Equal(t, "test", s)
			return &noOpCloser{outBuf}, nil
		}),
	)

	synctest.Test(t, func(t *testing.T) {
		err := interp.Evaluate(input)
		require.NoError(t, err)

		output := outBuf.String()
		expected := "echo 1\necho 2\n"
		assert.Equal(t, expected, output)
	})
}

func TestInterpretSingleCommands(t *testing.T) {
	tt := []struct {
		input        string
		expectedArgs []string
	}{
		{
			input:        " one   two ",
			expectedArgs: []string{"one", "two"},
		},
		{
			input:        "'one   two'",
			expectedArgs: []string{"one   two"},
		},
		{
			input:        "'one''two'",
			expectedArgs: []string{"onetwo"},
		},
		{
			input:        "one''two",
			expectedArgs: []string{"onetwo"},
		},
		{
			input:        `"one   two"`,
			expectedArgs: []string{"one   two"},
		},
		{
			input:        `"one""two"`,
			expectedArgs: []string{"onetwo"},
		},
		{
			input:        `"one" "two"`,
			expectedArgs: []string{"one", "two"},
		},
		{
			input:        `"one's two"`,
			expectedArgs: []string{"one's two"},
		},
		{
			input:        `three\ \ \ spaces`,
			expectedArgs: []string{"three   spaces"},
		},
		{
			input:        `one\   two`,
			expectedArgs: []string{"one ", "two"},
		},
		{
			input:        `one\ntwo`,
			expectedArgs: []string{"onentwo"},
		},
		{
			input:        `one\\two`,
			expectedArgs: []string{`one\two`},
		},
		{
			input:        `echo \'one\'`,
			expectedArgs: []string{`echo`, `'one'`},
		},
		{
			input:        `one \'\"two three\"\'`,
			expectedArgs: []string{`one`, `'"two`, `three"'`},
		},
		{
			input:        `one "A \\ escapes itself"`,
			expectedArgs: []string{`one`, `A \ escapes itself`},
		},
		{
			input:        `"A \" inside double quotes"`,
			expectedArgs: []string{`A " inside double quotes`},
		},
		{
			input:        `echo "example\"insidequotes"script\"`,
			expectedArgs: []string{"echo", `example"insidequotesscript"`},
		},
		{
			input:        `cat "/tmp/dog/'f  \53'"`,
			expectedArgs: []string{`cat`, `/tmp/dog/'f  \53'`},
		},
		{
			input:        `echo $HOME "Welcome ${HOME}."`,
			expectedArgs: []string{`echo`, `<HOME>`, `Welcome <HOME>.`},
		},
		{
			input:        `cat /tmp/owl/"meep mino"`,
			expectedArgs: []string{`cat`, `/tmp/owl/meep mino`},
		},
	}

	for _, test := range tt {
		t.Run(test.input, func(t *testing.T) {
			interp := NewInterpreter(
				WithIO(strings.NewReader(test.input), io.Discard, io.Discard),
				WithEnvFunc(func(s string) string {
					return "<HOME>"
				}),
				WithCmdLookupFunc(func(name string) (cmd CmdFunc, found bool, err error) {
					return func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
						assert.EqualValues(t, test.expectedArgs, args)
						return nil
					}, true, nil
				}),
			)

			synctest.Test(t, func(t *testing.T) {
				err := interp.Evaluate(test.input)
				require.NoError(t, err)
			})
		})
	}
}

type noOpCloser struct {
	io.ReadWriter
}

func (*noOpCloser) Close() error {
	return nil
}
