package plugin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

var _ shell.ShellPlugin = (*LuaPluginLoader)(nil)

type LuaPluginLoader struct {
	pluginDir string
	s         *shell.Shell
}

func NewLuaPluginLoader(pluginDir string) *LuaPluginLoader {
	return &LuaPluginLoader{
		pluginDir: pluginDir,
	}
}

func (*LuaPluginLoader) Name() string {
	return "Lua Plugin Loader"
}

func (l *LuaPluginLoader) Register(s *shell.Shell) {
	l.s = s
	s.AddHook(shell.HookInitialized, l.onInitialized)
}

func (l *LuaPluginLoader) onInitialized() {
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		l.s.Error(err.Error())
		return
	}

	plugins := []shell.ShellPlugin{}
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if ext != ".lua" {
			continue
		}

		pluginName := strings.TrimSuffix(entry.Name(), ext)
		pluginPath := filepath.Join(l.pluginDir, entry.Name())
		plugins = append(plugins, NewLuaPlugin(pluginName, pluginPath))
	}

	l.s.WithPlugins(plugins...)
	for _, plugin := range plugins {
		plugin.Register(l.s)
	}
}
