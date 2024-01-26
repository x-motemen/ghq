package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func doRm(c *cli.Context) error {
	var (
		name = c.Args().First()
		dry  = c.Bool("dry-run")
		w    = c.App.Writer
	)

	if name == "" {
		return fmt.Errorf("repository name is required")
	}

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
	if ok {
		return fmt.Errorf("directory %q does not exist", p)
	}

	if dry {
		fmt.Fprintf(w, "Would remove %s\n", p)
		return nil
	}

	ok, err = confirm(fmt.Sprintf("Remove %s?", p))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("aborted")
	}

	if err := os.RemoveAll(p); err != nil {
		return err
	}

	fmt.Fprintf(w, "Removed %s\n", p)
	return nil
}

func confirm(message string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, err
	}
	return response == "y", nil
}
