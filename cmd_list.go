package main

import (
	"context"
	"fmt"
	"io"
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
		printTree        = cmd.Bool("tree")
		bare             = cmd.Bool("bare")
	)

	if printTree && (printFullPaths || printUniquePaths) {
		return fmt.Errorf("--tree cannot be used with --full-path or --unique")
	}

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

	if printTree {
		repoList := make([]string, 0, len(repos))
		for _, repo := range repos {
			repoList = append(repoList, repo.RelPath)
		}
		sort.Strings(repoList)
		printRepoTree(w, repoList)
		return nil
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

type trieNode struct {
	children map[string]*trieNode
	keys     []string
}

func (n *trieNode) add(parts []string) {
	cur := n
	for _, p := range parts {
		if cur.children[p] == nil {
			if cur.children == nil {
				cur.children = map[string]*trieNode{}
			}
			cur.children[p] = &trieNode{}
			cur.keys = append(cur.keys, p)
		}
		cur = cur.children[p]
	}
}

func printRepoTree(w io.Writer, repos []string) {
	if len(repos) == 0 {
		fmt.Fprintln(w, "0 repositories")
		return
	}
	root := &trieNode{}
	for _, r := range repos {
		root.add(strings.Split(r, "/"))
	}
	fmt.Fprintln(w, ".")
	printTrieChildren(w, root, "")
	fmt.Fprintf(w, "\n%d repositories\n", len(repos))
}

func printTrieChildren(w io.Writer, n *trieNode, prefix string) {
	sort.Strings(n.keys)
	for i, key := range n.keys {
		child := n.children[key]
		// Collapse single-child intermediate nodes: golang.org/x instead of golang.org → x
		parts := []string{key}
		for len(child.keys) == 1 {
			parts = append(parts, child.keys[0])
			child = child.children[child.keys[0]]
		}
		label := strings.Join(parts, "/")
		last := i == len(n.keys)-1
		connector := "├── "
		childPrefix := "│   "
		if last {
			connector = "└── "
			childPrefix = "    "
		}
		fmt.Fprintf(w, "%s%s%s\n", prefix, connector, label)
		if len(child.keys) > 0 {
			printTrieChildren(w, child, prefix+childPrefix)
		}
	}
}
