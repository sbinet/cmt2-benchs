package cmt

import (
	"fmt"
	"os"
)

var g_verbose bool = false

func SetOutputLevel(verbose bool) {
	g_verbose = verbose
}

func debug(format string, args ...interface{}) (int, error) {
	if g_verbose {
		return fmt.Printf(format, args...)
	}
	return 0, nil
}

func path_exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// EOF
