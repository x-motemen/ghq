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
	Clone func(*vcsGetOption) error
	// Updates a cloned local repository.
	Update func(*vcsGetOption) error
	// Returns VCS specific files
	Contents func() []string
}

type vcsGetOption struct {
	url                        *url.URL
	dir                        string
	recursive, shallow, silent bool
	branch                     string
}

// GitBackend is the VCSBackend of git
var GitBackend = &VCSBackend{
	// support submodules?
	Clone: func(vg *vcsGetOption) error {
		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"clone"}
		if vg.shallow {
			args = append(args, "--depth", "1")
		}
		if vg.branch != "" {
			args = append(args, "--branch", vg.branch, "--single-branch")
		}
		args = append(args, vg.url.String(), vg.dir)

		return run(vg.silent)("git", args...)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "git", "pull", "--ff-only")
	},
	Contents: func() []string {
		return []string{".git"}
	},
}

// SubversionBackend is the VCSBackend for subversion
var SubversionBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"checkout"}
		if vg.shallow {
			args = append(args, "--depth", "1")
		}
		remote := vg.url
		if vg.branch != "" {
			copied := *vg.url
			remote = &copied
			remote.Path += "/branches/" + url.PathEscape(vg.branch)
		}
		args = append(args, remote.String(), vg.dir)

		return run(vg.silent)("svn", args...)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "svn", "update")
	},
	Contents: func() []string {
		return []string{".svn"}
	},
}

// GitsvnBackend is the VCSBackend for git-svn
var GitsvnBackend = &VCSBackend{
	// git-svn seems not supporting shallow clone currently.
	Clone: func(vg *vcsGetOption) error {
		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		remote := vg.url
		if vg.branch != "" {
			copied := *remote
			remote = &copied
			remote.Path += "/branches/" + url.PathEscape(vg.branch)
		}

		return run(vg.silent)("git", "svn", "clone", remote.String(), vg.dir)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "git", "svn", "rebase")
	},
	Contents: func() []string {
		return []string{".git/svn"}
	},
}

// MercurialBackend is the VCSBackend for mercurial
var MercurialBackend = &VCSBackend{
	// Mercurial seems not supporting shallow clone currently.
	Clone: func(vg *vcsGetOption) error {
		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		args := []string{"clone"}
		if vg.branch != "" {
			args = append(args, "--branch", vg.branch)
		}
		args = append(args, vg.url.String(), vg.dir)

		return run(vg.silent)("hg", args...)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "hg", "pull", "--update")
	},
	Contents: func() []string {
		return []string{".hg"}
	},
}

// DarcsBackend is the VCSBackend for darcs
var DarcsBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		if vg.branch != "" {
			return errors.New("Darcs does not support branch")
		}

		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"get"}
		if vg.shallow {
			args = append(args, "--lazy")
		}
		args = append(args, vg.url.String(), vg.dir)

		return run(vg.silent)("darcs", args...)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "darcs", "pull")
	},
	Contents: func() []string {
		return []string{"_darcs"}
	},
}

var cvsDummyBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		return errors.New("CVS clone is not supported")
	},
	Update: func(vg *vcsGetOption) error {
		return errors.New("CVS update is not supported")
	},
	Contents: func() []string {
		return []string{"CVS/Repository"}
	},
}

const fossilRepoName = ".fossil" // same as Go

// FossilBackend is the VCSBackend for fossil
var FossilBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		if vg.branch != "" {
			return errors.New("Fossil does not support cloning specific branch")
		}
		if err := os.MkdirAll(vg.dir, 0755); err != nil {
			return err
		}

		if err := run(vg.silent)("fossil", "clone", vg.url.String(), filepath.Join(vg.dir, fossilRepoName)); err != nil {
			return err
		}
		return runInDir(vg.silent)(vg.dir, "fossil", "open", fossilRepoName)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "fossil", "update")
	},
	Contents: func() []string {
		return []string{".fslckout", "_FOSSIL_"}
	},
}

// BazaarBackend is the VCSBackend for bazaar
var BazaarBackend = &VCSBackend{
	// bazaar seems not supporting shallow clone currently.
	Clone: func(vg *vcsGetOption) error {
		if vg.branch != "" {
			return errors.New("--branch option is unavailable for Bazaar since branch is included in remote URL")
		}
		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		return run(vg.silent)("bzr", "branch", vg.url.String(), vg.dir)
	},
	Update: func(vg *vcsGetOption) error {
		// Without --overwrite bzr will not pull tags that changed.
		return runInDir(vg.silent)(vg.dir, "bzr", "pull", "--overwrite")
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
