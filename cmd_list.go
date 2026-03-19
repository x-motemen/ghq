package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/urfave/cli/v3"
)

func doList(ctx context.Context, cmd *cli.Command) error {
	var (
		w                = cmd.Root().Writer
		query            = cmd.Args().First()
		exact            = cmd.Bool("exact")
		vcsBackend       = cmd.String("vcs")
		printFullPaths   = cmd.Bool("full-path")
		printUniquePaths = cmd.Bool("unique")
		bare             = cmd.Bool("bare")
	)

	filterByQuery := func(_ *LocalRepository) bool {
		return true
	}
	if query != "" {
		if hasSchemePattern.MatchString(query) || scpLikeURLPattern.MatchString(query) {
			if url, err := newURL(query, false, false); err == nil {
				if repo, err := LocalRepositoryFromURL(url, bare); err == nil {
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
			// Using smartcase searching
			if strings.ToLower(query) == query {
				filterByQuery = func(repo *LocalRepository) bool {
					return strings.Contains(strings.ToLower(repo.NonHostPath()), query) &&
						(host == "" || repo.PathParts[0] == host)
				}
			} else {
				filterByQuery = func(repo *LocalRepository) bool {
					return strings.Contains(repo.NonHostPath(), query) &&
						(host == "" || repo.PathParts[0] == host)
				}
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

	repoList := make([]string, 0, len(repos))
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
					repoList = append(repoList, p)
					break
				}
			}
		}
	} else {
		for _, repo := range repos {
			if printFullPaths {
				repoList = append(repoList, repo.FullPath)
			} else {
				repoList = append(repoList, repo.RelPath)
			}
		}
	}
	sort.Strings(repoList)
	for _, r := range repoList {
		fmt.Fprintln(w, r)
	}
	return nil
}
