package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

func doList(c *cli.Context) error {
	var (
		w                = c.App.Writer
		query            = c.Args().First()
		exact            = c.Bool("exact")
		ignoreCase       = c.Bool("ignore-case")
		vcsBackend       = c.String("vcs")
		printFullPaths   = c.Bool("full-path")
		printUniquePaths = c.Bool("unique")
	)

	if exact && ignoreCase {
		return fmt.Errorf("options -e(--exact) and -i(--ignore-case) are mutually exclusive")
	}

	filterByQuery := func(_ *LocalRepository) bool {
		return true
	}
	if query != "" {
		if hasSchemePattern.MatchString(query) || scpLikeURLPattern.MatchString(query) {
			if url, err := newURL(query, false, false); err == nil {
				if repo, err := LocalRepositoryFromURL(url); err == nil {
					query = filepath.ToSlash(repo.RelPath)
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
				p := repo.NonHostPath()
				q := query
				if ignoreCase {
					p = strings.ToLower(p)
					q = strings.ToLower(q)
				}
				return strings.Contains(p, q) &&
					(host == "" || repo.PathParts[0] == host)
			}
		}
	}

	var (
		repos []*LocalRepository
		mu    sync.Mutex
	)
	if err := walkLocalRepositories(vcsBackend, func(repo *LocalRepository) {
		if !filterByQuery(repo) {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		repos = append(repos, repo)
	}); err != nil {
		return fmt.Errorf("failed to filter repos while walkLocalRepositories(repo): %w", err)
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
