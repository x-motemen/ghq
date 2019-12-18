package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func doRoot(c *cli.Context) error {
	var (
		w   = c.App.Writer
		all = c.Bool("all")
	)
	if all {
		roots, err := localRepositoryRoots()
		if err != nil {
			return err
		}
		for _, root := range roots {
			fmt.Fprintln(w, root)
		}
		return nil
	}
	root, err := primaryLocalRepositoryRoot()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, root)
	return nil
}
