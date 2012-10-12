package cmt

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"
)

type Project struct {
	Mode Mode
	Name string
	Pkgs []*Pkg
	Uses []*Project
}

func NewProject(mode Mode, name string, npkgs int) *Project {
	proj := &Project{
		Mode: mode,
		Name: name,
		Pkgs: make([]*Pkg, npkgs),
		Uses: make([]*Project, 0),
	}
	for i := 0; i < npkgs; i++ {
		prefix := ""
		if rand.Int31n(2) < 1 {
			prefix = filepath.Join("Pre","Fix")
		}
		pkgname := fmt.Sprintf("Pkg_%04d", i+1)
		proj.Pkgs[i] = NewPkg(mode, name, prefix, pkgname)
	}
	return proj
}

func (proj *Project) generate() error {
	var err error
	err = proj.gen_packages()
	if err != nil {
		return err
	}
	return err
}

func (proj *Project) cleanup() error {
	debug("> rmdir [%s]\n", proj.Name)
	if path_exists(proj.Name) {
		return os.RemoveAll(proj.Name)
	}
	return nil
}

func (proj *Project) gen_structure() error {
	var err error
	for _, path := range []string{
		proj.Name,
		filepath.Join(proj.Name, "build"),
		filepath.Join(proj.Name, "cmt"),
	}{
		debug("> mkdir [%s]\n", path)
		err = os.MkdirAll(path, 0770)
		if err != nil {
			return err
		}
	}

	for _, pkg := range proj.Pkgs {
		err = pkg.gen_structure()
		if err != nil {
			return err
		}
	}

	return err
}

func (proj *Project) gen_packages() error {
	var err error
	for _, pkg := range proj.Pkgs {
		err = pkg.generate()
		if err != nil {
			return err
		}
	}
	return err
}

func (proj *Project) gen_config_file() error {
	n := filepath.Join(proj.Name, "CMakeLists.txt")
	debug("> gen [%s]\n", n)
	f, err := os.Create(n)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("cmake-proj").Parse(
`## {{.Name}}
cmake_minimum_required(VERSION 2.8)
include($ENV{CMTROOT}/cmake/CMTLib.cmake)
#-----------------
cmt_project({{.Name}} "")

{{with .Uses}}{{range .}}cmt_use_project({{.Name}})
{{end}}{{end}}

{{with .Pkgs}}{{range .}}cmt_has_package({{.Name}})
{{end}}{{end}}

`))
	err = tmpl.Execute(f, proj)
	if err != nil {
		return err
	}
	for _, pkg := range proj.Pkgs {
		err = pkg.gen_config_file()
		if err != nil {
			return err
		}
	}
	return err
}
// EOF
