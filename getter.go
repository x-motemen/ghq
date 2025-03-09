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

type getInfo struct {
	localRepository *LocalRepository
}

type getter struct {
	update, shallow, silent, ssh, recursive, bare bool
	vcs, branch, partial                          string
}

func (g *getter) get(argURL string) (getInfo, error) {
	u, err := newURL(argURL, g.ssh, false)
	if err != nil {
		return getInfo{}, fmt.Errorf("could not parse URL %q: %w", argURL, err)
	}
	branch := g.branch
	if pos := strings.LastIndexByte(u.Path, '@'); pos >= 0 {
		u.Path, branch = u.Path[:pos], u.Path[pos+1:]
	}
	remote, err := NewRemoteRepository(u)
	if err != nil {
		return getInfo{}, err
	}

	return g.getRemoteRepository(remote, branch)
}

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func (g *getter) getRemoteRepository(remote RemoteRepository, branch string) (getInfo, error) {
	remoteURL := remote.URL()
	local, err := LocalRepositoryFromURL(remoteURL, g.bare)
	if err != nil {
		return getInfo{}, err
	}
	info := getInfo{
		localRepository: local,
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
			return getInfo{}, err
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
				return getInfo{}, err
			}
		}
		if l := detectLocalRepoRoot(remoteURL.Path, repoURL.Path); l != "" {
			localRepoRoot = filepath.Join(local.RootPath, remoteURL.Hostname(), l)
		}

		if g.bare {
			localRepoRoot = localRepoRoot + ".git"
		}

		if remoteURL.Scheme == "codecommit" {
			repoURL, _ = url.Parse(remoteURL.Opaque)
		}
		if getRepoLock(localRepoRoot) {
			return info,
				vcs.Clone(&vcsGetOption{
					url:       repoURL,
					dir:       localRepoRoot,
					shallow:   g.shallow,
					silent:    g.silent,
					branch:    branch,
					recursive: g.recursive,
					bare:      g.bare,
					partial:   g.partial,
				})
		}
		return info, nil
	case g.update:
		logger.Log("update", fpath)
		vcs, localRepoRoot := local.VCS()
		if vcs == nil {
			return getInfo{}, fmt.Errorf("failed to detect VCS for %q", fpath)
		}
		repoURL := remoteURL
		if remoteURL.Scheme == "codecommit" {
			repoURL, _ = url.Parse(remoteURL.Opaque)
		}
		if getRepoLock(localRepoRoot) {
			return info, vcs.Update(&vcsGetOption{
				url:       repoURL,
				dir:       localRepoRoot,
				silent:    g.silent,
				recursive: g.recursive,
				bare:      g.bare,
			})
		}
		return info, nil
	}
	logger.Log("exists", fpath)
	return info, nil
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
