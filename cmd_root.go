package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func doRoot(c *cli.Context) error {
	roots, err := localRepositoryRoots(true)
	if err != nil {
		return err
	}
	if !c.Bool("all") {
		roots = roots[:1] // only the first root is needed
	}

	for _, root := range roots {
		_, err := fmt.Fprintln(c.App.Writer, root)
		if err != nil {
			return err
		}
	}

	return nil
}
