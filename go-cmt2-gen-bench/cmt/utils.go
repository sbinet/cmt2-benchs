package cmt

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
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

func CmtRoot() string {
	const dirname = "github.com/sbinet/cmt2-benchmarks/go-cmt2-gen-bench/cmt"
	for _, srcdir := range build.Default.SrcDirs() {
		dir := filepath.Join(srcdir, dirname)
		if path_exists(dir) {
			return dir
		}
	}
	panic("no CMTROOT available")
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
