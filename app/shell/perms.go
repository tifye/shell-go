package shell

import (
	"io/fs"
	"runtime"
)

func hasExecPerms(mode fs.FileMode) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return mode&0111 != 0
}
