package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

func doCreate(ctx context.Context, cmd *cli.Command) error {
	var (
		name = cmd.Args().First()
		vcs  = cmd.String("vcs")
		w    = cmd.Root().Writer
		bare = cmd.Bool("bare")
	)
	ignoreHost, err := ignoreHostFromCommand(cmd)
	if err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("repository name is required")
	}

	u, err := newURL(name, false, true)
	if err != nil {
		return err
	}

	localRepo, err := localRepositoryForURL(u, bare, ignoreHost)
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
		return err
	}

	vcsBackend, ok := vcsRegistry[vcs]
	if !ok {
		vcsBackend, _, err = remoteRepo.VCS()
		if err != nil {
			return err
		}
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
