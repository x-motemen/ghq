package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/x-motemen/ghq/logger"
)

var seen sync.Map

func getRepoLock(localRepoRoot string) bool {
	_, loaded := seen.LoadOrStore(localRepoRoot, struct{}{})
	return !loaded
}

type getter struct {
	update, shallow, silent, ssh, recursive, bare bool
	vcs, branch                                   string
}

func (g *getter) get(argURL string) error {
	u, err := newURL(argURL, g.ssh, false)
	if err != nil {
		return fmt.Errorf("could not parse URL %q: %w", argURL, err)
	}
	branch := g.branch
	if pos := strings.LastIndexByte(u.Path, '@'); pos >= 0 {
		u.Path, branch = u.Path[:pos], u.Path[pos+1:]
	}
	remote, err := NewRemoteRepository(u)
	if err != nil {
		return err
	}

	return g.getRemoteRepository(remote, branch)
}

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func (g *getter) getRemoteRepository(remote RemoteRepository, branch string) error {
	remoteURL := remote.URL()
	local, err := LocalRepositoryFromURL(remoteURL)
	if err != nil {
		return err
	}

	var (
		fpath   = local.FullPath
		newPath = false
	)

	_, err = os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			newPath = true
			err = nil
		}
		if err != nil {
			return err
		}
	}

	switch {
	case newPath:
		if remoteURL.Scheme == "codecommit" {
			logger.Log("clone", fmt.Sprintf("%s -> %s", remoteURL.Opaque, fpath))
		} else {
			logger.Log("clone", fmt.Sprintf("%s -> %s", remoteURL, fpath))
		}
		var (
			localRepoRoot = fpath
			repoURL       = remoteURL
		)
		vcs, ok := vcsRegistry[g.vcs]
		if !ok {
			vcs, repoURL, err = remote.VCS()
			if err != nil {
				return err
			}
		}
		if l := detectLocalRepoRoot(remoteURL.Path, repoURL.Path); l != "" {
			localRepoRoot = filepath.Join(local.RootPath, remoteURL.Hostname(), l)
		}

		if remoteURL.Scheme == "codecommit" {
			repoURL, _ = url.Parse(remoteURL.Opaque)
		}
		if getRepoLock(localRepoRoot) {
			return vcs.Clone(&vcsGetOption{
				url:       repoURL,
				dir:       localRepoRoot,
				shallow:   g.shallow,
				silent:    g.silent,
				branch:    branch,
				recursive: g.recursive,
				bare:      g.bare,
			})
		}
		return nil
	case g.update:
		logger.Log("update", fpath)
		vcs, localRepoRoot := local.VCS()
		if vcs == nil {
			return fmt.Errorf("failed to detect VCS for %q", fpath)
		}
		if getRepoLock(localRepoRoot) {
			return vcs.Update(&vcsGetOption{
				dir:       localRepoRoot,
				silent:    g.silent,
				recursive: g.recursive,
			})
		}
		return nil
	}
	logger.Log("exists", fpath)
	return nil
}

func detectLocalRepoRoot(remotePath, repoPath string) string {
	remotePath = strings.TrimSuffix(strings.TrimSuffix(remotePath, "/"), ".git")
	repoPath = strings.TrimSuffix(strings.TrimSuffix(repoPath, "/"), ".git")
	pathParts := strings.Split(repoPath, "/")
	pathParts = pathParts[1:]
	for i := 0; i < len(pathParts); i++ {
		subPath := "/" + path.Join(pathParts[i:]...)
		if subIdx := strings.Index(remotePath, subPath); subIdx >= 0 {
			return remotePath[0:subIdx] + subPath
		}
	}
	return ""
}
