package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/motemen/ghq/logger"
	"golang.org/x/xerrors"
)

type getter struct {
	update, shallow, silent, ssh bool
	vcs                          string
}

func (g *getter) get(argURL string) error {
	// If argURL is a "./foo" or "../bar" form,
	// find repository name trailing after github.com/USER/.
	parts := strings.Split(argURL, string(filepath.Separator))
	if parts[0] == "." || parts[0] == ".." {
		if wd, err := os.Getwd(); err == nil {
			path := filepath.Clean(filepath.Join(wd, filepath.Join(parts...)))

			var repoPath string
			roots, err := localRepositoryRoots()
			if err != nil {
				return err
			}
			for _, r := range roots {
				p := strings.TrimPrefix(path, r+string(filepath.Separator))
				if p != path && (repoPath == "" || len(p) < len(repoPath)) {
					repoPath = p
				}
			}

			if repoPath != "" {
				// Guess it
				logger.Log("resolved", fmt.Sprintf("relative %q to %q", argURL, "https://"+repoPath))
				argURL = "https://" + repoPath
			}
		}
	}

	u, err := newURL(argURL)
	if err != nil {
		return xerrors.Errorf("Could not parse URL %q: %w", argURL, err)
	}

	if g.ssh {
		// Assume Git repository if `-p` is given.
		if u, err = convertGitURLHTTPToSSH(u); err != nil {
			return xerrors.Errorf("Could not convet URL %q: %w", u, err)
		}
	}

	remote, err := NewRemoteRepository(u)
	if err != nil {
		return err
	}

	if remote.IsValid() == false {
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

	path := local.FullPath
	newPath := false

	_, err = os.Stat(path)
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
		logger.Log("clone", fmt.Sprintf("%s -> %s", remoteURL, path))

		vcs := vcsRegistry[g.vcs]
		repoURL := remoteURL
		if vcs == nil {
			vcs, repoURL = remote.VCS()
			if vcs == nil {
				return fmt.Errorf("Could not find version control system: %s", remoteURL)
			}
		}

		err := vcs.Clone(repoURL, path, g.shallow, g.silent)
		if err != nil {
			return err
		}
	} else {
		if g.update {
			logger.Log("update", path)
			vcs, repoPath := local.VCS()
			if vcs == nil {
				return fmt.Errorf("failed to detect VCS for %q", path)
			}
			vcs.Update(repoPath, g.silent)
		} else {
			logger.Log("exists", path)
		}
	}
	return nil
}
