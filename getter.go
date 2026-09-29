package main

import (
	"context"
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
	update, shallow, silent, ssh, recursive bool
	bareMode                                BareMode
	vcs, branch, partial                    string
}

func (g *getter) get(ctx context.Context, argURL string) (getInfo, error) {
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

	return g.getRemoteRepository(ctx, remote, branch)
}

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func (g *getter) getRemoteRepository(ctx context.Context, remote RemoteRepository, branch string) (getInfo, error) {
	remoteURL := remote.URL()
	local, err := LocalRepositoryFromURL(remoteURL, g.bareMode)
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

		// localRepoRoot at this point does NOT include the bare-mode
		// suffix (detectLocalRepoRoot strips ".git" if present); the
		// switch below adds it (either as a suffix on the leaf directory
		// or as a ".git" subdirectory).
		switch g.bareMode {
		case BareClean:
			// Clean-bare: the bare gitdir lives inside the leaf directory
			// as ".git" (repo/.git), so we join rather than concatenate.
			localRepoRoot = filepath.Join(localRepoRoot, ".git")
		case BareClassic:
			// Classic bare: the leaf directory itself is the bare gitdir
			// (repo -> repo.git), so we suffix the leaf and deliberately
			// do NOT use filepath.Join, which would create a new segment.
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
					bare:      g.bareMode != BareNone,
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
				bare:      g.bareMode != BareNone,
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
		if before, _, ok := strings.Cut(remotePath, subPath); ok {
			return before + subPath
		}
	}
	return ""
}
