package cmd

type Command struct {
	Name string
	Run  func(input []byte) error
}
