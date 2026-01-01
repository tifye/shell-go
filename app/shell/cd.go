package shell

import (
	"errors"
	"fmt"
	"os"

	"github.com/codecrafters-io/shell-starter-go/app/cmd"
)

func NewCDCommandFunc(s *Shell) cmd.CommandFunc {
	return func() *cmd.Command {
		return &cmd.Command{
			Name: "cmd",
			Run: func(cmd *cmd.Command, args []string) error {
				if len(args) < 2 {
					return nil
				}

				target := args[1]
				if !os.IsPathSeparator(target[0]) {
					target = s.WorkingDir + target
				}
				target, err := s.FullPathFunc(target)
				if err != nil {
					return fmt.Errorf("failed to get full path of %q: %w\n", target, err)
				}

				f, err := s.FS.OpenFile(target, os.O_RDONLY)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						_, err := fmt.Fprintf(cmd.Stdout, "cd: %s: No such file or directory\n", args[1])
						return err
					}

					return fmt.Errorf("failed to check target location: %s\n", err)
				}
				_ = f.Close()

				s.WorkingDir = target

				return nil
			},
		}
	}
}
