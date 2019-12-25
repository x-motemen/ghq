package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func doGet(c *cli.Context) error {
	var (
		argURL  = c.Args().Get(0)
		andLook = c.Bool("look")
	)
	g := &getter{
		update:    c.Bool("update"),
		shallow:   c.Bool("shallow"),
		ssh:       c.Bool("p"),
		vcs:       c.String("vcs"),
		silent:    c.Bool("silent"),
		branch:    c.String("branch"),
		recursive: !c.Bool("no-recursive"),
	}

	if argURL == "" {
		return fmt.Errorf("no project args specified. see `ghq get -h` for more details")
	}

	if err := g.get(argURL); err != nil {
		return err
	}
	if andLook {
		return doLook(c)
	}
	return nil
}
