package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sbinet/cmt2-benchmarks/go-cmt2-gen-bench/cmt"
)

var g_mode = flag.String("mode", "cmake", "generation mode")
var g_projects = flag.Int("nprojs", 1, "number of projects to generate")
var g_packages = flag.Int("npkgs", 5, "number of packages in each project")
var g_uses = flag.String("uses", "", "comma-separated list of uses-packages")
var g_verbose = flag.Bool("verbose", false, "")

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

func main() {
	flag.Parse()
	fmt.Printf(":: %s...\n", os.Args[0])
	fmt.Printf(":: mode:     %q\n", *g_mode)
	fmt.Printf(":: projects: %d\n", *g_projects)
	fmt.Printf(":: packages: %d\n", *g_packages)

	uses := []string{}
	for _, v := range strings.Split(*g_uses, ",") {
		if v != "" {
			uses = append(uses, v)
		}
	}
	fmt.Printf(":: uses:     %v\n", uses)

	cmt.SetOutputLevel(*g_verbose)
	mode := cmt.Mode(*g_mode)
	dir := "test_"+*g_mode

	var err error

	gen, err := cmt.NewGenerator(mode, dir, *g_projects, *g_packages, uses)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf(">> generate project(s)...\n")
	err = gen.Generate()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf(">> run build...\n")
	err = gen.Run()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf(":: bye.\n")
}

// EOF
