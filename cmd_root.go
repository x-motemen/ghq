package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func doRoot(ctx context.Context, cmd *cli.Command) error {
	roots, err := localRepositoryRoots(true)
	if err != nil {
		return err
	}
	if !cmd.Bool("all") {
		roots = roots[:1] // only the first root is needed
	}

	for _, root := range roots {
		_, err := fmt.Fprintln(cmd.Root().Writer, root)
		if err != nil {
			return err
		}
	}

	return nil
}
