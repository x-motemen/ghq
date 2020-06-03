package main

import (
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

func doCreate(c *cli.Context) error {
	var (
		name = c.Args().First()
		vcs  = c.String("vcs")
		w    = c.App.Writer
	)
	u, err := newURL(name, false, true)
	if err != nil {
		return err
	}

	localRepo, err := LocalRepositoryFromURL(u)
	if err != nil {
		return err
	}

	p := localRepo.FullPath
	ok, err := isNotExistOrEmpty(p)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("directory %q already exists and not empty", p)
	}

	remoteRepo, err := NewRemoteRepository(u)
	if err != nil {
		return nil
	}

	vcsBackend, ok := vcsRegistry[vcs]
	if !ok {
		vcsBackend, u = remoteRepo.VCS()
	}
	if vcsBackend == nil {
		return fmt.Errorf("failed to init: unsupported VCS")
	}

	initFunc := vcsBackend.Init
	if initFunc == nil {
		return fmt.Errorf("failed to init: unsupported VCS")
	}

	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}

	if err := initFunc(p); err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, p)
	return err
}

func isNotExistOrEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
