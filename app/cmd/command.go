package cmd

type Command struct {
	Name string
	Run  func(args []string) error
}
