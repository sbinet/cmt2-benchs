package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var g_mode = flag.String("mode", "cmake", "generation mode")
var g_projects = flag.Int("nprojs", 1, "number of projects to generate")
var g_packages = flag.Int("npkgs", 5, "number of packages in each project")
var g_uses = flag.String("uses", "", "comma-separated list of uses-packages")
var g_verbose = flag.Bool("verbose", false, "")

var hdr_tmpl = template.Must(template.New("hdr").Parse(
	`/* -*- c++ -*- */
#ifndef LIB_{{.Name}}_H
#define LIB_{{.Name}}_H 1

{{with .Uses}}{{range .}}#include "{{.Name}}/Lib{{.Name}}.h"
{{end}}{{end}}

#ifdef _MSC_VER
# define API_EXPORT __declspec( dllexport )
#else
#if __GNUC__ >= 4
# define API_EXPORT __attribute__((visibility("default")))
#else
# define API_EXPORT
#endif
#endif

class API_EXPORT C{{.Name}}
{
public:
   C{{.Name}}();
   ~C{{.Name}}();
   void f();
private:
{{with .Uses}}{{range .}} C{{.Name}} m_{{.Name}};
{{end}}{{end}}
};
#endif /* !LIB_{{.Name}}_H */
/* EOF */
`))

var cxx_tmpl = template.Must(template.New("cxx").Parse(
`// Lib{{.Name}}.cxx
#include <iostream>
#include "{{.Name}}/Lib{{.Name}}.h"

C{{.Name}}::C{{.Name}}()
{
   std::cout << ":: c-tor C{{.Name}}\n";
}

C{{.Name}}::~C{{.Name}}()
{
   std::cout << ":: d-tor C{{.Name}}\n";
}

void
C{{.Name}}::f()
{
   std::cout << ":: C{{.Name}}.f\n";
   {{with .Uses}}{{range .}}m_{{.Name}}.f();
   {{end}}{{end}}
}
`))

func debug(format string, args ...interface{}) (int, error) {
	if *g_verbose {
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

type mode_t string

const (
	cmt2waf   mode_t = "cmt2waf"
	cmt2cmake mode_t = "cmt2cmake"
	cmake     mode_t = "cmake"
	waf       mode_t = "waf"
)

type CmtPkg struct {
	Mode     mode_t
	Project  string
	Prefix   string
	Name     string
	FullName string
	Path     string
	Uses     []*CmtPkg
}

func NewCmtPkg(mode mode_t, project, prefix, pkg string) *CmtPkg {
	fullname := filepath.Join(prefix, pkg)
	return &CmtPkg{
		Mode:     mode,
		Project:  project,
		Prefix:   prefix,
		Name:     pkg,
		FullName: fullname,
		Path:     filepath.Join(project, fullname),
		Uses:     make([]*CmtPkg, 0),
	}
}

func (pkg *CmtPkg) cleanup() error {
	debug("> rmdir [%s]\n", pkg.Path)
	if path_exists(pkg.Path) {
		return os.RemoveAll(pkg.Path)
	}
	return nil
}

func (pkg *CmtPkg) gen_structure() error {
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

func (pkg *CmtPkg) gen_headers() error {
	n := filepath.Join(pkg.Path, pkg.Name, fmt.Sprintf("Lib%s.h", pkg.Name))
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	return hdr_tmpl.Execute(f, pkg)
}

func (pkg *CmtPkg) gen_sources() error {
	n := filepath.Join(pkg.Path, "src", fmt.Sprintf("Lib%s.cxx", pkg.Name))
	f, err := os.Create(n)
	if err != nil {
		return err
	}
	return cxx_tmpl.Execute(f, pkg)
}

func (pkg *CmtPkg) gen_test() error {
	n := filepath.Join(pkg.Path, "src", fmt.Sprintf("test%s.cxx", pkg.Name))
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

func (pkg *CmtPkg) gen_requirements() error {
	n := filepath.Join(pkg.Path, "cmt", "requirements")
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

func (pkg *CmtPkg) gen() error {
	var err error
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

type CmtProject struct {
	Mode mode_t
	Name string
	Pkgs []*CmtPkg
	Uses []*CmtProject
}

func NewCmtProject(mode mode_t, name string, npkgs int) *CmtProject {
	proj := &CmtProject{
		Mode: mode,
		Name: name,
		Pkgs: make([]*CmtPkg, npkgs),
		Uses: make([]*CmtProject, 0),
	}
	for i := 0; i < npkgs; i++ {
		prefix := ""
		if rand.Int31n(2) < 1 {
			prefix = filepath.Join("Pre","Fix")
		}
		pkgname := fmt.Sprintf("Pkg_%04d", i+1)
		proj.Pkgs[i] = NewCmtPkg(mode, name, prefix, pkgname)
	}
	return proj
}

func (proj *CmtProject) gen() error {
	return nil
}

func (proj *CmtProject) cleanup() error {
	debug("> rmdir [%s]\n", proj.Name)
	if path_exists(proj.Name) {
		return os.RemoveAll(proj.Name)
	}
	return nil
}

func (proj *CmtProject) gen_structure() error {
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

func (proj *CmtProject) gen_packages() error {
	var err error
	
	return err
}

// CmtGenerator constructs a CMT project with a hierarchy of projects, each
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
type CmtGenerator struct {
	Mode     mode_t
	Projects []*CmtProject
	NPkgs    int
	Uses     []string
}

func NewGenerator(mode mode_t, nprojects, npkgs int, uses []string) *CmtGenerator {
	gen := &CmtGenerator{
		Mode: mode,
		Projects: make([]*CmtProject, nprojects),
		NPkgs: npkgs,
		Uses: uses,
	}
	for i := 0; i < nprojects; i++ {
		name := fmt.Sprintf("Proj_%04d", i)
		npkgs := rand.Intn(npkgs)+1
		debug(":: creating project [%s]...\n", name)
		gen.Projects[i] = NewCmtProject(mode, name, npkgs)
	}
	return gen
}

func (gen *CmtGenerator) cleanup_projects() error {
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

func (gen *CmtGenerator) gen_structure() error {
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

func (gen *CmtGenerator) generate() error {
	var err error
	
	return err
}

func (gen *CmtGenerator) gen_config_files() error {
	var err error
	
	return err
}

func (gen *CmtGenerator) Run() error {
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

	if path_exists("test") {
		err := os.RemoveAll("test")
		if err != nil {
			panic(err.Error())
		}
	}
	err := os.Mkdir("test", 0700)
	if err != nil {
		panic(err.Error())
	}

	err = os.Chdir("test")
	if err != nil {
		panic(err.Error())
	}

	gen := NewGenerator(mode_t(*g_mode), *g_projects, *g_packages, uses)
	err = gen.Run()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf(":: bye.\n")
}

// EOF
