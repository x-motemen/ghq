package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Songmu/gitconfig"
	"github.com/urfave/cli/v3"
)

const configIgnoreHost = "ghq.ignoreHost"

type repoNaming struct {
	ignoreHost       bool
	hostfulRelPath   string
	hostlessRelPath  string
	canonicalRelPath string
	hostfulParts     []string
	hostlessParts    []string
	canonicalParts   []string
}

func ignoreHostFromCommand(cmd *cli.Command) (bool, error) {
	if cmd.Bool("ignore-host") {
		return true, nil
	}

	ignoreHost, err := gitconfig.Bool(configIgnoreHost)
	if err != nil {
		if gitconfig.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return ignoreHost, nil
}

func newRepoNaming(remoteURL *url.URL, bare, ignoreHost bool) repoNaming {
	repoParts := splitRepoPath(remoteURL.Path)
	if len(repoParts) > 0 {
		repoParts[len(repoParts)-1] = strings.TrimSuffix(repoParts[len(repoParts)-1], ".git")
	}

	hostlessParts := append([]string(nil), repoParts...)
	hostfulParts := append([]string{remoteURL.Hostname()}, repoParts...)
	if bare {
		if len(hostlessParts) > 0 {
			hostlessParts[len(hostlessParts)-1] = hostlessParts[len(hostlessParts)-1] + ".git"
		}
		if len(hostfulParts) > 0 {
			hostfulParts[len(hostfulParts)-1] = hostfulParts[len(hostfulParts)-1] + ".git"
		}
	}

	canonicalParts := hostfulParts
	canonicalRelPath := relPathFromParts(hostfulParts)
	if ignoreHost {
		canonicalParts = hostlessParts
		canonicalRelPath = relPathFromParts(hostlessParts)
	}

	return repoNaming{
		ignoreHost:       ignoreHost,
		hostfulRelPath:   relPathFromParts(hostfulParts),
		hostlessRelPath:  relPathFromParts(hostlessParts),
		canonicalRelPath: canonicalRelPath,
		hostfulParts:     hostfulParts,
		hostlessParts:    hostlessParts,
		canonicalParts:   canonicalParts,
	}
}

func splitRepoPath(urlPath string) []string {
	trimmed := strings.Trim(urlPath, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func relPathFromParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return filepath.ToSlash(filepath.Join(parts...))
}

func repoNamingFromURL(remoteURL *url.URL, bare, ignoreHost bool) repoNaming {
	return newRepoNaming(remoteURL, bare, ignoreHost)
}

func newLocalRepositoryFromNaming(remoteURL *url.URL, naming repoNaming, root string) *LocalRepository {
	return &LocalRepository{
		FullPath:  filepath.Join(root, naming.canonicalRelPath),
		RelPath:   naming.canonicalRelPath,
		RootPath:  root,
		PathParts: append([]string(nil), naming.canonicalParts...),
	}
}

func localRepositoryForURL(remoteURL *url.URL, bare, ignoreHost bool) (*LocalRepository, error) {
	naming := repoNamingFromURL(remoteURL, bare, ignoreHost)
	match, err := findCanonicalRepositoryMatch(remoteURL, naming)
	if err != nil {
		return nil, err
	}
	if match != nil {
		return match, nil
	}

	remoteURLStr := remoteURL.String()
	if remoteURL.Scheme == "codecommit" {
		remoteURLStr = remoteURL.Opaque
	}
	root, err := getRoot(remoteURLStr)
	if err != nil {
		return nil, err
	}
	return newLocalRepositoryFromNaming(remoteURL, naming, root), nil
}

func lookupLocalRepositoryForURL(remoteURL *url.URL, bare, ignoreHost bool) (*LocalRepository, error) {
	naming := repoNamingFromURL(remoteURL, bare, ignoreHost)
	if !ignoreHost {
		match, err := findCanonicalRepositoryMatch(remoteURL, naming)
		if err != nil {
			return nil, err
		}
		if match != nil {
			return match, nil
		}
		remoteURLStr := remoteURL.String()
		if remoteURL.Scheme == "codecommit" {
			remoteURLStr = remoteURL.Opaque
		}
		root, err := getRoot(remoteURLStr)
		if err != nil {
			return nil, err
		}
		return newLocalRepositoryFromNaming(remoteURL, naming, root), nil
	}

	var (
		exact   *LocalRepository
		matches []*LocalRepository
	)
	if err := walkAllLocalRepositories(func(repo *LocalRepository) {
		if repo.HostlessRelPath() != naming.hostlessRelPath {
			return
		}
		if repo.RelPath == naming.canonicalRelPath {
			exact = repo
		}
		matches = append(matches, repo)
	}); err != nil {
		return nil, err
	}

	if exact != nil {
		ok, err := repoMatchesRemoteHost(exact, remoteURL)
		if err == nil && !ok {
			return nil, ignoreHostConflictError(naming.hostlessRelPath, matches)
		}
		if len(matches) == 1 {
			return exact, nil
		}
	}

	switch len(matches) {
	case 0:
		remoteURLStr := remoteURL.String()
		if remoteURL.Scheme == "codecommit" {
			remoteURLStr = remoteURL.Opaque
		}
		root, err := getRoot(remoteURLStr)
		if err != nil {
			return nil, err
		}
		return newLocalRepositoryFromNaming(remoteURL, naming, root), nil
	case 1:
		return matches[0], nil
	default:
		return nil, ignoreHostConflictError(naming.hostlessRelPath, matches)
	}
}

func findCanonicalRepositoryMatch(remoteURL *url.URL, naming repoNaming) (*LocalRepository, error) {
	var (
		exact      *LocalRepository
		collisions []*LocalRepository
	)

	if err := walkAllLocalRepositories(func(repo *LocalRepository) {
		if repo.RelPath == naming.canonicalRelPath {
			exact = repo
		}
		if naming.ignoreHost && repo.HostlessRelPath() == naming.hostlessRelPath && repo.RelPath != naming.canonicalRelPath {
			collisions = append(collisions, repo)
		}
	}); err != nil {
		return nil, err
	}

	if !naming.ignoreHost {
		return exact, nil
	}

	if exact != nil {
		ok, err := repoMatchesRemoteHost(exact, remoteURL)
		if err == nil {
			if !ok {
				return nil, ignoreHostConflictError(naming.hostlessRelPath, []*LocalRepository{exact})
			}
		} else if len(collisions) == 0 {
			return exact, nil
		}
	}

	if len(collisions) > 0 {
		return nil, ignoreHostConflictError(naming.hostlessRelPath, collisions)
	}

	return exact, nil
}

func repoMatchesRemoteHost(repo *LocalRepository, remoteURL *url.URL) (bool, error) {
	vcs, repoPath := repo.VCS()
	if vcs == nil || vcs.RemoteURL == nil {
		return false, fmt.Errorf("cannot determine remote host for %q", repo.RelPath)
	}

	remote, err := vcs.RemoteURL(repoPath)
	if err != nil {
		return false, err
	}

	u, err := newURL(remote, false, false)
	if err != nil {
		return false, err
	}

	return u.Hostname() == remoteURL.Hostname(), nil
}

func ignoreHostConflictError(hostlessRelPath string, repos []*LocalRepository) error {
	paths := make([]string, 0, len(repos))
	for _, repo := range repos {
		paths = append(paths, repo.RelPath)
	}
	return fmt.Errorf("ignore-host naming collision for %q; existing repositories use the same owner/repo path: %s. Migrate or remove the conflicting repository first", hostlessRelPath, strings.Join(paths, ", "))
}
