package cmt

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

func (proj *Project) gen() error {
	return nil
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
	
	return err
}


// EOF
