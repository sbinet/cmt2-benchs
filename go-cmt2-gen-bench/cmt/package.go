package cmt

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type Pkg struct {
	Mode     Mode
	Project  string
	Prefix   string
	Name     string
	FullName string
	Path     string
	Uses     []*Pkg
}

func NewPkg(mode Mode, project, prefix, pkg string) *Pkg {
	fullname := filepath.Join(prefix, pkg)
	return &Pkg{
		Mode:     mode,
		Project:  project,
		Prefix:   prefix,
		Name:     pkg,
		FullName: fullname,
		Path:     filepath.Join(project, fullname),
		Uses:     make([]*Pkg, 0),
	}
}

func (pkg *Pkg) cleanup() error {
	debug("> rmdir [%s]\n", pkg.Path)
	if path_exists(pkg.Path) {
		return os.RemoveAll(pkg.Path)
	}
	return nil
}

func (pkg *Pkg) gen_structure() error {
	for _, path := range []string{
		pkg.Path,
		filepath.Join(pkg.Path, pkg.Name),
		filepath.Join(pkg.Path, "src"),
		filepath.Join(pkg.Path, "cmt"),
	} {
		debug("> mkdir [%s]\n", path)
		err := os.MkdirAll(path, 0770)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pkg *Pkg) gen_headers() error {
	n := filepath.Join(pkg.Path, pkg.Name, fmt.Sprintf("Lib%s.h", pkg.Name))
	debug("> gen [%s]\n", n)
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	return hdr_tmpl.Execute(f, pkg)
}

func (pkg *Pkg) gen_sources() error {
	n := filepath.Join(pkg.Path, "src", fmt.Sprintf("Lib%s.cxx", pkg.Name))
	debug("> gen [%s]\n", n)
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	return cxx_tmpl.Execute(f, pkg)
}

func (pkg *Pkg) gen_test() error {
	n := filepath.Join(pkg.Path, "src", fmt.Sprintf("test%s.cxx", pkg.Name))
	debug("> gen [%s]\n", n)
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("test").Parse(
`// test{{.Name}}.cxx
#include <iostream>
#include "{{.Name}}/Lib{{.Name}}.h"

int main()
{
  C{{.Name}} o;
  o.f();
  return 0;
}
// EOF
`))
	return tmpl.Execute(f, pkg)
}

func (pkg *Pkg) gen_requirements() error {
	n := filepath.Join(pkg.Path, "cmt", "requirements")
	debug("> gen [%s]\n", n)
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("cmt-req").Parse(
		`#---------------------------------
package {{.Name}}

# package deps
{{with .Uses}}{{range .}}use {{.Name}} {{.Name}}* {{.Prefix}}
{{end}}{{end}}

# constituents
macro Lib{{.Name}}_linkopts "{{with .Uses}}{{range .}}Lib{{.Name}} {{end}}{{end}}"
macro test{{.Name}}_linkopts "{{with .Uses}}{{range .}}Lib{{.Name}} {{end}}{{end}}"
library Lib{{.Name}} Lib{{.Name}}.cxx

program test{{.Name}} test{{.Name}}.cxx
#---------------------------------
## EOF ##
`))
	return tmpl.Execute(f, pkg)
}

func (pkg *Pkg) generate() error {
	var err error
	debug("> gen [%s]\n", pkg.FullName)
	err = pkg.gen_headers()
	if err != nil {
		return err
	}
	err = pkg.gen_sources()
	if err != nil {
		return err
	}
	err = pkg.gen_test()
	if err != nil {
		return err
	}
	err = pkg.gen_requirements()
	if err != nil {
		return err
	}
	return err
}

func (pkg *Pkg) gen_config_file() error {
	n := filepath.Join(pkg.Path, "CMakeLists.txt")
	debug("> gen [%s]\n", n)
	tmpl := template.Must(template.New("cmake_pkg").Parse(
`## {{.Name}}
cmake_minimum_required(VERSION 2.8)
include($ENV{CMTROOT}/cmake/CMTLib.cmake)
#-----------------
cmt_package({{.Name}})

{{with .Uses}}{{range .}}cmt_use_package ({{.Name}})
{{end}}{{end}}

cmt_library(Lib{{.Name}} src/Lib{{.Name}}.cxx "{{with .Uses}}{{range .}}cmt_use_package (Lib{{.Name}}){{end}}{{end}}")
cmt_executable(test{{.Name}} src/test{{.Name}}.cxx Lib{{.Name}})
cmt_test(mytest{{.Name}})

#-----------------

cmt_action ()

## EOF ##
`))

	f, err := os.Create(n)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, pkg)
}

// EOF
