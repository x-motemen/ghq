package main

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/motemen/ghq/cmdutil"
)

func run(silent bool) func(command string, args ...string) error {
	if silent {
		return cmdutil.RunSilently
	}
	return cmdutil.Run
}

func runInDir(silent bool) func(dir, command string, args ...string) error {
	if silent {
		return cmdutil.RunInDirSilently
	}
	return cmdutil.RunInDir
}

// A VCSBackend represents a VCS backend.
type VCSBackend struct {
	// Clones a remote repository to local path.
	Clone func(*url.URL, string, bool, bool, string) error
	// Updates a cloned local repository.
	Update func(string, bool) error
	// Returns VCS specific files
	Contents func() []string
}

// GitBackend is the VCSBackend of git
var GitBackend = &VCSBackend{
	// support submodules?
	Clone: func(remote *url.URL, local string, shallow, silent bool, branch string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"clone"}
		if shallow {
			args = append(args, "--depth", "1")
		}
		if branch != "" {
			args = append(args, "--branch", branch, "--single-branch")
		}
		args = append(args, remote.String(), local)

		return run(silent)("git", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "git", "pull", "--ff-only")
	},
	Contents: func() []string {
		return []string{".git"}
	},
}

// SubversionBackend is the VCSBackend for subversion
var SubversionBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool, branch string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"checkout"}
		if shallow {
			args = append(args, "--depth", "1")
		}
		if branch != "" {
			copied := *remote
			remote = &copied
			remote.Path += "/branches/" + url.PathEscape(branch)
		}
		args = append(args, remote.String(), local)

		return run(silent)("svn", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "svn", "update")
	},
	Contents: func() []string {
		return []string{".svn"}
	},
}

// GitsvnBackend is the VCSBackend for git-svn
var GitsvnBackend = &VCSBackend{
	// git-svn seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool, branch string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		if branch != "" {
			copied := *remote
			remote = &copied
			remote.Path += "/branches/" + url.PathEscape(branch)
		}

		return run(silent)("git", "svn", "clone", remote.String(), local)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "git", "svn", "rebase")
	},
	Contents: func() []string {
		return []string{".git/svn"}
	},
}

// MercurialBackend is the VCSBackend for mercurial
var MercurialBackend = &VCSBackend{
	// Mercurial seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool, branch string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		args := []string{"clone"}
		if branch != "" {
			args = append(args, "--branch", branch)
		}
		args = append(args, remote.String(), local)

		return run(silent)("hg", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "hg", "pull", "--update")
	},
	Contents: func() []string {
		return []string{".hg"}
	},
}

// DarcsBackend is the VCSBackend for darcs
var DarcsBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool, branch string) error {
		if branch != "" {
			return errors.New("Darcs does not support branch")
		}

		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"get"}
		if shallow {
			args = append(args, "--lazy")
		}
		args = append(args, remote.String(), local)

		return run(silent)("darcs", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "darcs", "pull")
	},
	Contents: func() []string {
		return []string{"_darcs"}
	},
}

var cvsDummyBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool, branch string) error {
		return errors.New("CVS clone is not supported")
	},
	Update: func(local string, silent bool) error {
		return errors.New("CVS update is not supported")
	},
	Contents: func() []string {
		return []string{"CVS/Repository"}
	},
}

const fossilRepoName = ".fossil" // same as Go

// FossilBackend is the VCSBackend for fossil
var FossilBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool, branch string) error {
		if branch != "" {
			return errors.New("Fossil does not support cloning specific branch")
		}
		if err := os.MkdirAll(local, 0755); err != nil {
			return err
		}

		if err := run(silent)("fossil", "clone", remote.String(), filepath.Join(local, fossilRepoName)); err != nil {
			return err
		}
		return runInDir(silent)(local, "fossil", "open", fossilRepoName)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "fossil", "update")
	},
	Contents: func() []string {
		return []string{".fslckout", "_FOSSIL_"}
	},
}

// BazaarBackend is the VCSBackend for bazaar
var BazaarBackend = &VCSBackend{
	// bazaar seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool, branch string) error {
		if branch != "" {
			return errors.New("--branch option is unavailable for Bazaar since branch is included in remote URL")
		}
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		return run(silent)("bzr", "branch", remote.String(), local)
	},
	Update: func(local string, silent bool) error {
		// Without --overwrite bzr will not pull tags that changed.
		return runInDir(silent)(local, "bzr", "pull", "--overwrite")
	},
	Contents: func() []string {
		return []string{".bzr"}
	},
}

var vcsRegistry = map[string]*VCSBackend{
	"git":        GitBackend,
	"github":     GitBackend,
	"svn":        SubversionBackend,
	"subversion": SubversionBackend,
	"git-svn":    GitsvnBackend,
	"hg":         MercurialBackend,
	"mercurial":  MercurialBackend,
	"darcs":      DarcsBackend,
	"fossil":     FossilBackend,
	"bzr":        BazaarBackend,
	"bazaar":     BazaarBackend,
}
