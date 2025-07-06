package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/x-motemen/ghq/cmdutil"
	"github.com/urfave/cli/v2"
)

func doCheck(c *cli.Context) error {
	var (
		w          = c.App.Writer
		vcsBackend = c.String("vcs")
	)

	var (
		repos []*LocalRepository
		mu    sync.Mutex
	)
	if err := walkLocalRepositories(vcsBackend, func(repo *LocalRepository) {
		mu.Lock()
		defer mu.Unlock()
		repos = append(repos, repo)
	}); err != nil {
		return fmt.Errorf("failed to get repositories: %w", err)
	}

	for _, repo := range repos {
		out, err := cmdutil.RunAndGetOutput("git", "-C", repo.FullPath, "status", "--porcelain")
		if err != nil {
			// Handle cases where the directory is not a git repository
			continue
		}
		stashes, _ := cmdutil.RunAndGetOutput("git", "-C", repo.FullPath, "stash", "list")

		if strings.TrimSpace(out) == "" && strings.TrimSpace(stashes) == "" {
			continue
		}

		fmt.Fprintln(w, repo.RelPath)
		if strings.TrimSpace(out) != "" {
			fmt.Fprintln(w, "  Uncommitted changes:")
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				fmt.Fprintln(w, "    "+line)
			}
		}
		if strings.TrimSpace(stashes) != "" {
			fmt.Fprintln(w, "  Stashes:")
			for _, line := range strings.Split(strings.TrimSpace(stashes), "\n") {
				fmt.Fprintln(w, "    "+line)
			}
		}
	}

	return nil
}
