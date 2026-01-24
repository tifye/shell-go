package shell

type hook func()

type hooks struct {
	preReadHooks []hook
}

func newHooks() *hooks {
	return &hooks{
		preReadHooks: make([]hook, 0),
	}
}

func (h *hooks) AddPreReadHook(f hook) {
	h.preReadHooks = append(h.preReadHooks, f)
}

func (h *hooks) runPreReadHooks() {
	for _, f := range h.preReadHooks {
		f()
	}
}
