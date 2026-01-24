package shell

import "github.com/codecrafters-io/shell-starter-go/app/shell/terminal"

type KeyHandler interface {
	Handle(t terminal.Item)
	Next(k KeyHandler)
}

type KeyHandlers struct {
	handlers map[terminal.ItemType]KeyHandler
}

func NewKeyHandlers() *KeyHandlers {
	return &KeyHandlers{
		handlers: map[terminal.ItemType]KeyHandler{},
	}
}

func (k *KeyHandlers) Handle(t terminal.Item) {
	handler, ok := k.handlers[t.Type]
	if ok {
		handler.Handle(t)
	}
}

func (k *KeyHandlers) Use(t terminal.ItemType, next KeyHandler) {
	handler, ok := k.handlers[t]
	if !ok {
		k.handlers[t] = next
	} else {
		handler.Next(next)
	}
}
