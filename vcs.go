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
	Init   func(dir string) error
	// Returns VCS specific files
	Contents []string
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
		if vg.recursive {
			args = append(args, "--recursive")
		}
		args = append(args, vg.url.String(), vg.dir)

		return run(vg.silent)("git", args...)
	},
	Update: func(vg *vcsGetOption) error {
		if _, err := os.Stat(filepath.Join(vg.dir, ".git/svn")); err == nil {
			return GitsvnBackend.Update(vg)
		}
		err := runInDir(vg.silent)(vg.dir, "git", "pull", "--ff-only")
		if err != nil {
			return err
		}
		if vg.recursive {
			return runInDir(vg.silent)(vg.dir, "git", "submodule", "update", "--init", "--recursive")
		}
		return nil
	},
	Init: func(dir string) error {
		return cmdutil.RunInDirStderr(dir, "git", "init")
	},
	Contents: []string{".git"},
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
			args = append(args, "--depth", "immediates")
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
	Contents: []string{".svn"},
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
	Contents: []string{".git/svn"},
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
	Init: func(dir string) error {
		return cmdutil.RunInDirStderr(dir, "hg", "init")
	},
	Contents: []string{".hg"},
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
	Init: func(dir string) error {
		return cmdutil.RunInDirStderr(dir, "darcs", "init")
	},
	Contents: []string{"_darcs"},
}

var cvsDummyBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		return errors.New("CVS clone is not supported")
	},
	Update: func(vg *vcsGetOption) error {
		return errors.New("CVS update is not supported")
	},
	Contents: []string{"CVS/Repository"},
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
	Init: func(dir string) error {
		if err := cmdutil.RunInDirStderr(dir, "fossil", "init", fossilRepoName); err != nil {
			return err
		}
		return cmdutil.RunInDirStderr(dir, "fossil", "open", fossilRepoName)
	},
	Contents: []string{".fslckout", "_FOSSIL_"},
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
	Init: func(dir string) error {
		return cmdutil.RunInDirStderr(dir, "bzr", "init")
	},
	Contents: []string{".bzr"},
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
