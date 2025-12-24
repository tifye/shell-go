package cmd

type CommandRunFunc func(args []string) error

type Command struct {
	Name string
	Run  CommandRunFunc
}
