package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/motemen/ghq/cmdutil"
	"github.com/urfave/cli"
)

func doCreate(c *cli.Context) error {
	var (
		name = c.Args().First()
		w    = c.App.Writer
	)
	u, err := newURL(name, false, true)
	if err != nil {
		return err
	}
	root, err := getRoot(u.String())
	if err != nil {
		return err
	}
	p := filepath.Join(root, u.Hostname(), u.Path)
	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Dir = p
	if err := cmdutil.RunCommand(cmd, true); err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, p)
	return err
}
