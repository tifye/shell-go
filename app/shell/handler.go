package shell

import "github.com/codecrafters-io/shell-starter-go/app/shell/terminal"

var defaultHandler KeyHandler = func(_ terminal.Item) error {
	return nil
}

type KeyHandler func(terminal.Item) error

type KeyMiddlewareFunc func(next KeyHandler) KeyHandler

type KeyHandlers struct {
	handlers map[terminal.ItemType][]KeyMiddlewareFunc
}

func newEventHandlers() *KeyHandlers {
	return &KeyHandlers{
		handlers: map[terminal.ItemType][]KeyMiddlewareFunc{},
	}
}

func (e *KeyHandlers) Use(typ terminal.ItemType, h KeyMiddlewareFunc) {
	e.handlers[typ] = append(e.handlers[typ], h)
}

func (e *KeyHandlers) handle(item terminal.Item) error {
	funcs, ok := e.handlers[item.Type]
	if !ok {
		return nil
	}

	handler := defaultHandler
	for _, middlewareFunc := range funcs {
		handler = middlewareFunc(handler)
	}

	return handler(item)
}
