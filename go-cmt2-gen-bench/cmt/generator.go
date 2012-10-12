package cmt

import (
	"fmt"
	"math/rand"
	"os"
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
}

func NewGenerator(mode Mode, nprojects, npkgs int, uses []string) *Generator {
	gen := &Generator{
		Mode:     mode,
		Projects: make([]*Project, nprojects),
		NPkgs:    npkgs,
		Uses:     uses,
	}
	for i := 0; i < nprojects; i++ {
		name := fmt.Sprintf("Proj_%04d", i)
		npkgs := rand.Intn(npkgs) + 1
		debug(":: creating project [%s]...\n", name)
		gen.Projects[i] = NewProject(mode, name, npkgs)
	}
	return gen
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
  set(CMTPROJECTPATH "${CMTROOT}/test")
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

// EOF
