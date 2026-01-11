package terminal

type ColorStack struct {
	colors [][]byte
}

func newColorStack() *ColorStack {
	return &ColorStack{
		colors: make([][]byte, 0),
	}
}

func (s *ColorStack) Push(c []byte) {
	s.colors = append(s.colors, c)
}

func (s *ColorStack) Pop() {
	if len(s.colors) > 0 {
		s.colors = s.colors[:len(s.colors)-1]
	}
}

func (s *ColorStack) Top() []byte {
	if len(s.colors) == 0 {
		return resetColor
	}
	return s.colors[len(s.colors)-1]
}
