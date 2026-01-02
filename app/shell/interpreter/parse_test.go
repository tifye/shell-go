package interpreter

import (
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
	"github.com/stretchr/testify/assert"
)

func TestParseSingleCommands(t *testing.T) {
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
			expectedArgs: []string{"onetwo"},
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
	}

	for _, test := range tt {
		t.Run(test.input, func(t *testing.T) {
			prog, err := Parse(test.input, CommandLookuperFunc(func(s string) (*cmd.Command, bool, error) {
				return &cmd.Command{
					Name: test.expectedArgs[0],
					Run:  assertCommandCall(t, test.expectedArgs),
				}, true, nil
			}), getEnv)
			if assert.NoError(t, err) {
				assert.Len(t, prog.cmds, 1)
			}

			err = prog.Run()
			assert.NoError(t, err)
		})
	}
}

func assertCommandCall(t *testing.T, expectedArgs []string) cmd.CommandRunFunc {
	t.Helper()
	return func(cmd *cmd.Command, args []string) error {
		assert.EqualValues(t, expectedArgs, args)
		return nil
	}
}

type CommandLookuperFunc func(string) (*cmd.Command, bool, error)

func (f CommandLookuperFunc) LookupCommand(name string) (*cmd.Command, bool, error) {
	return f(name)
}

func getEnv(n string) string {
	return "<" + n + ">"
}
