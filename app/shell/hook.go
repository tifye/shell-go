package shell

import "github.com/codecrafters-io/shell-starter-go/assert"

//go:generate stringer -type Hook -trimprefix Hook
type Hook int

const (
	HookPreRead Hook = iota
	HookPreExit
	// HookInitialized runs after the shell is initialized but before
	// it begins its Read Print Eval loop
	HookInitialized
)

type HookFunc func()

type hooks struct {
	triggerMap map[Hook][]HookFunc
}

func newHooks() *hooks {
	return &hooks{
		triggerMap: map[Hook][]HookFunc{},
	}
}

func (h *hooks) runHooks(trigger Hook) {
	funcs, ok := h.triggerMap[trigger]
	if !ok {
		return
	}

	for _, f := range funcs {
		f()
	}
}

func (h *hooks) AddHook(trigger Hook, hook func()) {
	assert.NotNil(hook)

	if _, ok := h.triggerMap[trigger]; !ok {
		h.triggerMap[trigger] = make([]HookFunc, 0, 1)
	}

	h.triggerMap[trigger] = append(h.triggerMap[trigger], hook)
}
