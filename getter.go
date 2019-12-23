package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/motemen/ghq/logger"
)

var (
	seen = make(map[string]bool)
	mu   = &sync.Mutex{}
)

func getRepoLock(localRepoRoot string) bool {
	mu.Lock()
	defer func() {
		seen[localRepoRoot] = true
		mu.Unlock()
	}()
	return !seen[localRepoRoot]
}

type getter struct {
	update, shallow, silent, ssh, recursive bool
	vcs, branch                             string
}

func (g *getter) get(argURL string) error {
	// If argURL is a "./foo" or "../bar" form,
	// find repository name trailing after github.com/USER/.
	argURL = filepath.ToSlash(argURL)
	parts := strings.Split(argURL, "/")
	if parts[0] == "." || parts[0] == ".." {
		if wd, err := os.Getwd(); err == nil {
			path := filepath.Clean(filepath.Join(wd, filepath.Join(parts...)))

			var localRepoRoot string
			roots, err := localRepositoryRoots(true)
			if err != nil {
				return err
			}
			for _, r := range roots {
				p := strings.TrimPrefix(path, r+string(filepath.Separator))
				if p != path && (localRepoRoot == "" || len(p) < len(localRepoRoot)) {
					localRepoRoot = filepath.ToSlash(p)
				}
			}

			if localRepoRoot != "" {
				// Guess it
				logger.Log("resolved", fmt.Sprintf("relative %q to %q", argURL, "https://"+localRepoRoot))
				argURL = "https://" + localRepoRoot
			}
		}
	}

	u, err := newURL(argURL)
	if err != nil {
		return fmt.Errorf("Could not parse URL %q: %w", argURL, err)
	}

	if g.ssh {
		// Assume Git repository if `-p` is given.
		if u, err = convertGitURLHTTPToSSH(u); err != nil {
			return fmt.Errorf("Could not convert URL %q: %w", u, err)
		}
	}

	remote, err := NewRemoteRepository(u)
	if err != nil {
		return err
	}

	if !remote.IsValid() {
		return fmt.Errorf("Not a valid repository: %s", u)
	}

	return g.getRemoteRepository(remote)
}

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func (g *getter) getRemoteRepository(remote RemoteRepository) error {
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

	if newPath {
		logger.Log("clone", fmt.Sprintf("%s -> %s", remoteURL, fpath))
		var (
			vcs           = vcsRegistry[g.vcs]
			localRepoRoot = fpath
			repoURL       = remoteURL
		)
		vcs2, repoURL2 := remote.VCS()
		if vcs == nil && vcs2 == nil {
			return fmt.Errorf("Could not find version control system: %s", remoteURL)
		}
		if vcs == nil {
			vcs = vcs2
		}
		if repoURL2 != nil {
			repoURL = repoURL2
		}
		l := detectLocalRepoRoot(
			strings.TrimSuffix(remoteURL.Path, ".git"),
			strings.TrimSuffix(repoURL.Path, ".git"))
		if l != "" {
			localRepoRoot = filepath.Join(local.RootPath, remoteURL.Hostname(), l)
		}

		if getRepoLock(localRepoRoot) {
			return vcs.Clone(&vcsGetOption{
				url:       repoURL,
				dir:       localRepoRoot,
				shallow:   g.shallow,
				silent:    g.silent,
				branch:    g.branch,
				recursive: g.recursive,
			})
		}
		return nil
	}
	if g.update {
		logger.Log("update", fpath)
		vcs, localRepoRoot := local.VCS()
		if vcs == nil {
			return fmt.Errorf("failed to detect VCS for %q", fpath)
		}
		if getRepoLock(localRepoRoot) {
			return vcs.Update(&vcsGetOption{
				dir:    localRepoRoot,
				silent: g.silent,
			})
		}
		return nil
	}
	logger.Log("exists", fpath)
	return nil
}

func detectLocalRepoRoot(remotePath, repoPath string) string {
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
