package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/motemen/ghq/logger"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func doImport(c *cli.Context) error {
	var parallel = c.Bool("parallel")
	g := &getter{
		update:  c.Bool("update"),
		shallow: c.Bool("shallow"),
		ssh:     c.Bool("p"),
		vcs:     c.String("vcs"),
		silent:  c.Bool("silent"),
	}
	if parallel {
		// force silent in parallel import
		g.silent = true
	}

	eg := &errgroup.Group{}
	sem := make(chan struct{}, 6)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if parallel {
			eg.Go(func() error {
				sem <- struct{}{}
				defer func() { <-sem }()
				if err := g.get(line); err != nil {
					logger.Log("error", err.Error())
				}
				return nil
			})
		} else {
			if err := g.get(line); err != nil {
				logger.Log("error", err.Error())
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("While reading input: %s", err)
	}
	return eg.Wait()
}
