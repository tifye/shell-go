package assert

import "fmt"

func NotNil(v any, msgs ...string) {
	if v == nil {
		if len(msgs) > 0 {
			panic(msgs[0])
		}
		panic("expected non-nil value")
	}
}

func Assert(c bool, msg ...string) {
	if !c {
		var m string
		if len(msg) > 0 {
			m = msg[0]
		} else {
			m = fmt.Sprintf("expected %t", c)
		}
		panic(m)
	}
}

func Assertf(c bool, format string, args ...any) {
	if !c {
		panic(fmt.Sprintf(format, args...))
	}
}
