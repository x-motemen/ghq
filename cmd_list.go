package main

import (
	"context"
	"fmt"
	"os"
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
				if repo, err := LocalRepositoryFromURL(url, bareModeFromClassicBool(bare)); err == nil {
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

	type scannedRepository struct {
		*LocalRepository
		info       os.FileInfo
		symlink    bool
		matches    bool
		identity   directoryIdentity
		identified bool
		duplicate  bool
	}
	var scanned []scannedRepository
	var mu sync.Mutex
	if err := walkLocalRepositories(vcsBackend, func(repo *LocalRepository, info os.FileInfo, symlink bool) {
		mu.Lock()
		defer mu.Unlock()
		scanned = append(scanned, scannedRepository{LocalRepository: repo, info: info, symlink: symlink})
	}); err != nil {
		return fmt.Errorf("failed to filter repos while walkLocalRepositories(repo): %w", err)
	}

	repos := make([]*scannedRepository, 0, len(scanned))
	needIdentities := false
	for i := range scanned {
		repo := &scanned[i]
		repo.matches = filterByQuery(repo.LocalRepository)
		if repo.matches {
			repos = append(repos, repo)
			needIdentities = needIdentities || repo.symlink
		}
	}
	if len(repos) == 0 {
		return nil
	}
	// A target's value records whether it also matches the query.
	targets := make(map[directoryIdentity]bool)
	if needIdentities {
		// Windows needs a handle to obtain an identity; bound those reads.
		slots := make(chan struct{}, 6)
		var workers sync.WaitGroup
		for i := range scanned {
			repo := &scanned[i]
			if repo.symlink && !repo.matches {
				continue
			}
			slots <- struct{}{}
			workers.Go(func() {
				defer func() { <-slots }()
				repo.identity, repo.identified = identifyDirectory(repo.FullPath, repo.info)
			})
		}
		workers.Wait()
		for _, repo := range scanned {
			if repo.identified && !repo.symlink {
				targets[repo.identity] = targets[repo.identity] || repo.matches
			}
		}
	}
	for _, repo := range repos {
		if repo.symlink && repo.identified {
			_, repo.duplicate = targets[repo.identity]
		}
	}

	repoList := make([]string, 0, len(repos))
	if printUniquePaths {
		subpathCount := map[string]int{} // Count duplicated subpaths (ex. foo/dotfiles and bar/dotfiles)
		reposCount := map[string]int{}   // Check duplicated repositories among roots
		// Keep aliases in collision counts so omitting them does not shorten other names.
		for _, repo := range repos {
			if reposCount[repo.RelPath] == 0 {
				for _, p := range repo.Subpaths() {
					subpathCount[p]++
				}
			}
			reposCount[repo.RelPath]++
		}
		uniqueName := func(repo *LocalRepository) string {
			for _, name := range repo.Subpaths() {
				if subpathCount[name] == 1 {
					return name
				}
			}
			return ""
		}
		represented := make(map[directoryIdentity]bool)
		for _, repo := range repos {
			if repo.duplicate || (reposCount[repo.RelPath] > 1 && !repo.IsUnderPrimaryRoot()) {
				continue
			}
			if name := uniqueName(repo.LocalRepository); name != "" {
				repoList = append(repoList, name)
				if repo.identified {
					represented[repo.identity] = true
				}
			}
		}
		// Retain an identifiable alias only when its canonical target has no usable name.
		fallbacks := make(map[directoryIdentity]string)
		for _, repo := range repos {
			if !repo.duplicate || !targets[repo.identity] || represented[repo.identity] ||
				(reposCount[repo.RelPath] > 1 && !repo.IsUnderPrimaryRoot()) {
				continue
			}
			name := uniqueName(repo.LocalRepository)
			previous := fallbacks[repo.identity]
			if name != "" && (previous == "" || strings.Count(name, "/") < strings.Count(previous, "/") ||
				(strings.Count(name, "/") == strings.Count(previous, "/") && name < previous)) {
				fallbacks[repo.identity] = name
			}
		}
		for _, name := range fallbacks {
			repoList = append(repoList, name)
		}
	} else {
		for _, repo := range repos {
			if repo.duplicate {
				continue
			}
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
