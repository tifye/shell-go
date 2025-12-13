package assert

func NotNil(v any) {
	if v == nil {
		panic("expected non-nil value")
	}
}
