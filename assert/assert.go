package assert

import "fmt"

func NotNil(v any) {
	if v == nil {
		panic("expected non-nil value")
	}
}

func Assert(c bool) {
	if !c {
		panic(fmt.Sprintf("expected %t", c))
	}
}
