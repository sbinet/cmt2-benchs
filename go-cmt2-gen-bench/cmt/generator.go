package cmt

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// Generator constructs a CMT project with a hierarchy of projects, each
// containing a hierarchy of packages, each of which containing:
//  - sources of a (C++) library (with one C++ class)
//  - sources of a test program (which instantiates one object of the class)
//  - a cmt requirements file
//  - or a CMakeLists.txt
//
// Projects may use other projects.
// Packages may use other packages.
// For each used package, the package's class includes the used classes, 
// and instantiates one object for each the used classe.
type Generator struct {
	Mode     Mode
	Projects []*Project
	NPkgs    int
	Uses     []string
	Dir      string
}

func NewGenerator(mode Mode, dir string, nprojects, npkgs int, uses []string) (*Generator, error) {
	var err error
	testdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	gen := &Generator{
		Mode:     mode,
		Projects: make([]*Project, nprojects),
		NPkgs:    npkgs,
		Uses:     uses,
		Dir: testdir,
	}

	if path_exists(testdir) {
		err := os.RemoveAll(testdir)
		if err != nil {
			return nil, err
		}
	}
	err = os.Mkdir(testdir, 0700)
	if err != nil {
		return nil, err
	}

	// FIXME: we shouldn't do that!!
	// instead, we should have all the paths relative to this testdir!
	err = os.Chdir(testdir)
	if err != nil {
		return nil, err
	}

	for i := 0; i < nprojects; i++ {
		name := fmt.Sprintf("Proj_%04d", i)
		npkgs := rand.Intn(npkgs) + 1
		debug(":: creating project [%s]...\n", name)
		gen.Projects[i] = NewProject(mode, name, npkgs)
	}
	return gen, nil
}

func (gen *Generator) cleanup_projects() error {
	for _, proj := range gen.Projects {
		if proj == nil {
			continue
		}
		err := proj.cleanup()
		if err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) gen_structure() error {
	debug("> gen_structure...\n")

	for _, proj := range gen.Projects {
		if proj == nil {
			continue
		}
		err := proj.gen_structure()
		if err != nil {
			return err
		}
	}

	if path_exists("build") {
		err := os.RemoveAll("build")
		if err != nil {
			return err
		}
	}

	tmpl := template.Must(template.New("cmakelist").Parse(
		`## CMakeLists.txt
cmake_minimum_required(VERSION 2.8)
include($ENV{CMTROOT}/cmake/CMTLib.cmake)
cmake_minimum_required(VERSION 2.8)

set(CMTROOT "$ENV{CMTROOT}")
set(CMTPROJECTPATH "$ENV{CMTPROJECTPATH}")
if("${CMTPROJECTPATH}" STREQUAL "")
  set(CMTPROJECTPATH "{{.Dir}}")
endif()

unset(status)
cmt_init(status)

if("${status}" STREQUAL "stop")
 return()
endif()

cmt_off()

cmt_use_project(work)

cmt_project(work "")

{{with .Projects}}{{range .}}cmt_use_project({{.Name}})
{{end}}{{end}}

cmt_action()

## EOF ##
`))

	f, err := os.Create("CMakeLists.txt")
	if err != nil {
		return err
	}

	return tmpl.Execute(f, gen)
}

func (gen *Generator) generate() error {
	debug("> generate...\n")
	var err error
	for _, proj := range gen.Projects {
		// if proj == nil {
		// 	continue
		// }
		err = proj.generate()
		if err != nil {
			return err
		}
	}
	return err
}

func (gen *Generator) gen_config_files() error {
	debug("> gen_config_files...\n")
	var err error
	for _, proj := range gen.Projects {
		err = proj.gen_config_file()
		if err != nil {
			return err
		}
	}
	return err
}

func (gen *Generator) Generate() error {
	var err error

	err = gen.cleanup_projects()
	if err != nil {
		return err
	}

	err = gen.gen_structure()
	if err != nil {
		return err
	}

	err = gen.generate()
	if err != nil {
		return err
	}

	err = gen.gen_config_files()
	if err != nil {
		return err
	}
	return err
}

func (gen *Generator) Run() error {
	var err error
	// setup environment...
	err = os.Setenv("CMTROOT", CmtRoot())
	if err != nil {
		return err
	}
	err = os.Setenv("CMTPROJECTPATH", gen.Dir)
	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	testdir := pwd

	bdir := "build"
	if path_exists(bdir) {
		err = os.RemoveAll(bdir)
		if err != nil {
			return err
		}
	}
	err = os.Mkdir(bdir, 0700)
	if err != nil {
		return err
	}

	bdir = filepath.Join(testdir, bdir)

	cmds := []*exec.Cmd{}

	// run the build
	switch gen.Mode {
	case CMake:
		cmd := exec.Command("cmake", 
			"--build=.", filepath.Join(testdir,"CMakeLists.txt"),
			)
		cmd.Dir = bdir
		if g_verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		cmds = append(cmds, cmd)

		cmd = exec.Command("make")
		cmd.Dir = testdir
		if g_verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		cmds = append(cmds, cmd)

	default:
		panic("mode [" + string(gen.Mode) + "] unknown!")
	}

	for _, cmd := range cmds {
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return err
}

// EOF
