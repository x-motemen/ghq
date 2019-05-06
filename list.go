package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

func doList(c *cli.Context) error {
	var (
		w                = c.App.Writer
		query            = c.Args().First()
		exact            = c.Bool("exact")
		vcsBackend       = c.String("vcs")
		printFullPaths   = c.Bool("full-path")
		printUniquePaths = c.Bool("unique")
	)

	filterByQuery := func(_ *LocalRepository) bool {
		return true
	}
	if query != "" {
		if hasSchemePattern.MatchString(query) || scpLikeURLPattern.MatchString(query) {
			if url, err := newURL(query); err == nil {
				if repo, err := LocalRepositoryFromURL(url); err == nil {
					query = repo.RelPath
				}
			}
		}

		if exact {
			filterByQuery = func(repo *LocalRepository) bool {
				return repo.Matches(query)
			}
		} else {
			var host string
			paths := strings.Split(query, "/")
			if len(paths) > 1 && looksLikeAuthorityPattern.MatchString(paths[0]) {
				query = strings.Join(paths[1:], "/")
				host = paths[0]
			}
			filterByQuery = func(repo *LocalRepository) bool {
				return strings.Contains(repo.NonHostPath(), query) &&
					(host == "" || repo.PathParts[0] == host)
			}
		}
	}
	filterByVCS := func(repo *LocalRepository) bool {
		if vcsBackend == "" {
			return true
		}
		vcs, _ := repo.VCS()
		return vcsRegistry[vcsBackend] == vcs
	}

	repos := []*LocalRepository{}
	if err := walkLocalRepositories(func(repo *LocalRepository) {
		if !filterByQuery(repo) || !filterByVCS(repo) {
			return
		}
		repos = append(repos, repo)
	}); err != nil {
		return xerrors.Errorf("failed to filter repos while walkLocalRepositories(repo): %w", err)
	}

	if printUniquePaths {
		subpathCount := map[string]int{} // Count duplicated subpaths (ex. foo/dotfiles and bar/dotfiles)
		reposCount := map[string]int{}   // Check duplicated repositories among roots

		// Primary first
		for _, repo := range repos {
			if reposCount[repo.RelPath] == 0 {
				for _, p := range repo.Subpaths() {
					subpathCount[p] = subpathCount[p] + 1
				}
			}

			reposCount[repo.RelPath] = reposCount[repo.RelPath] + 1
		}

		for _, repo := range repos {
			if reposCount[repo.RelPath] > 1 && !repo.IsUnderPrimaryRoot() {
				continue
			}

			for _, p := range repo.Subpaths() {
				if subpathCount[p] == 1 {
					fmt.Fprintln(w, p)
					break
				}
			}
		}
	} else {
		for _, repo := range repos {
			if printFullPaths {
				fmt.Fprintln(w, repo.FullPath)
			} else {
				fmt.Fprintln(w, repo.RelPath)
			}
		}
	}
	return nil
}
