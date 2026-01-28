package plugin

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"github.com/codecrafters-io/shell-starter-go/app/shell/terminal"
	lua "github.com/yuin/gopher-lua"
)

var _ shell.ShellPlugin = (*LuaPlugin)(nil)

type LuaPlugin struct {
	name     string
	filename string
	lstate   *lua.LState

	s  *shell.Shell
	tw *terminal.TermWriter
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
	l.tw = s.Terminal().Writer()

	l.lstate.PreloadModule("shell", l.shellLoader)
	l.lstate.PreloadModule("shellplugin", l.shellPluginLoader)

	if err := l.lstate.DoFile(l.filename); err != nil {
		fmt.Println(err)
	}

	s.AddHook(shell.HookPreExit, func() {
		l.lstate.Close()
	})
}

func (l *LuaPlugin) shellPluginLoader(lstate *lua.LState) int {
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
		if _, err := luaCall[*lua.LNilType](lstate, lfunc); err != nil {
			l.s.Error(err.Error())
		}
	})

	return 0
}

func (l *LuaPlugin) SetPromptStringFunc(lstate *lua.LState) int {
	lfunc := lstate.ToFunction(1)

	l.s.Terminal().PromptStringFunc = func() string {
		val, err := luaCall[lua.LString](lstate, lfunc)
		if err != nil {
			l.s.Error(err.Error())
			return ""
		}
		return val.String()
	}
	return 0
}

func (l *LuaPlugin) shellLoader(lstate *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"StagePushForegroundColor": l.StagePushForegroundColor,
		"StagePopForegroundColor":  l.StagePopForegroundColor,
		"StageString":              l.StageString,
	}

	mod := lstate.SetFuncs(lstate.NewTable(), exports)

	lstate.Push(mod)
	return 1
}

func (l *LuaPlugin) StagePushForegroundColor(lstate *lua.LState) int {
	lcolor := lstate.ToString(1)
	color := []byte(lcolor)
	l.tw.StagePushForegroundColor(color)
	return 0
}

func (l *LuaPlugin) StagePopForegroundColor(lstate *lua.LState) int {
	l.tw.StagePopForegroundColor()
	return 0
}

func (l *LuaPlugin) StageString(lstate *lua.LState) int {
	lstr := lstate.ToString(1)
	l.tw.StageString(lstr)
	return 0
}

func luaCall[R lua.LValue](lstate *lua.LState, lfunc *lua.LFunction, args ...lua.LValue) (R, error) {
	numRet := 0
	var aux R
	switch any(aux).(type) {
	case lua.LString, lua.LNumber, lua.LBool, lua.LFunction,
		lua.LChannel, lua.LTable, lua.LState, lua.LUserData:
		numRet = 1
	case *lua.LNilType:
	default:
		return aux, errors.New("invalid lua type for return")
	}

	err := lstate.CallByParam(lua.P{
		Fn:      lfunc,
		NRet:    numRet,
		Protect: true,
	}, args...)
	if err != nil {
		return aux, err
	}

	if numRet == 0 {
		return aux, err
	}

	ret := lstate.Get(-1)
	defer lstate.Pop(1)

	if retVal, ok := ret.(R); ok {
		return retVal, nil
	} else {
		return aux, fmt.Errorf("expected return type of %q but got %q", reflect.TypeOf(aux).String(), reflect.TypeOf(ret).String())
	}
}
