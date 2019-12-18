package main

import (
	"os"

	"github.com/urfave/cli"
)

func doGet(c *cli.Context) error {
	var (
		argURL  = c.Args().Get(0)
		andLook = c.Bool("look")
	)
	g := &getter{
		update:  c.Bool("update"),
		shallow: c.Bool("shallow"),
		ssh:     c.Bool("p"),
		vcs:     c.String("vcs"),
		silent:  c.Bool("silent"),
		branch:  c.String("branch"),
	}

	if argURL == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	if err := g.get(argURL); err != nil {
		return err
	}
	if andLook {
		return doLook(c)
	}
	return nil
}
