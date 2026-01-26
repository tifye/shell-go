package plugin

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
	lua "github.com/yuin/gopher-lua"
)

var _ shell.ShellPlugin = (*LuaPlugin)(nil)

type LuaPlugin struct {
	name     string
	filename string
	lstate   *lua.LState

	s *shell.Shell
}

func NewLuaPlugin(name, filename string) *LuaPlugin {
	lstate := lua.NewState(lua.Options{
		IncludeGoStackTrace: true,
	})
	return &LuaPlugin{
		name:     name,
		filename: filename,
		lstate:   lstate,
	}
}

func (l *LuaPlugin) Name() string {
	return l.name
}

func (l *LuaPlugin) Register(s *shell.Shell) {
	l.s = s

	l.lstate.PreloadModule("shellplugin", l.loader)

	if err := l.lstate.DoFile(l.filename); err != nil {
		fmt.Println(err)
	}

	s.AddHook(shell.HookPreExit, func() {
		l.lstate.Close()
	})
}

func (l *LuaPlugin) loader(lstate *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"SetPromptStringFunc": l.SetPromptStringFunc,
		"AddHook":             l.AddHook,
	}

	mod := lstate.SetFuncs(lstate.NewTable(), exports)

	lstate.Push(mod)
	return 1
}

func (l *LuaPlugin) AddHook(lstate *lua.LState) int {
	lhook := lstate.ToString(1)
	lfunc := lstate.ToFunction(2)

	l.s.AddHook(shell.Hook(lhook), func() {
		err := lstate.CallByParam(lua.P{
			Fn:      lfunc,
			NRet:    0,
			Protect: true,
		})
		if err != nil {
			l.s.Error(err.Error())
			return
		}
	})

	return 0
}

func (l *LuaPlugin) SetPromptStringFunc(lstate *lua.LState) int {
	lfunc := lstate.ToFunction(1)

	l.s.Terminal().PromptStringFunc = func() string {
		err := lstate.CallByParam(lua.P{
			Fn:      lfunc,
			NRet:    1,
			Protect: true,
		})
		if err != nil {
			l.s.Error(err.Error())
			return ""
		}

		val := lstate.Get(-1)
		defer lstate.Pop(1)
		return lua.LVAsString(val)
	}
	return 0
}
