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
	Clone func(*url.URL, string, bool, bool) error
	// Updates a cloned local repository.
	Update func(string, bool) error
}

var GitBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"clone"}
		if shallow {
			args = append(args, "--depth", "1")
		}
		args = append(args, remote.String(), local)

		return run(silent)("git", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "git", "pull", "--ff-only")
	},
}

var SubversionBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		args := []string{"checkout"}
		if shallow {
			args = append(args, "--depth", "1")
		}
		args = append(args, remote.String(), local)

		return run(silent)("svn", args...)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "svn", "update")
	},
}

var GitsvnBackend = &VCSBackend{
	// git-svn seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return run(silent)("git", "svn", "clone", remote.String(), local)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "git", "svn", "rebase")
	},
}

var MercurialBackend = &VCSBackend{
	// Mercurial seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return run(silent)("hg", "clone", remote.String(), local)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "hg", "pull", "--update")
	},
}

var DarcsBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool) error {
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
}

var cvsDummyBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool) error {
		return errors.New("CVS clone is not supported")
	},
	Update: func(local string, silent bool) error {
		return errors.New("CVS update is not supported")
	},
}

const fossilRepoName = ".fossil" // same as Go

var FossilBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string, shallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		err = run(silent)("fossil", "clone", remote.String(), filepath.Join(dir, fossilRepoName))
		if err != nil {
			return err
		}

		err = os.Chdir(dir)
		if err != nil {
			return err
		}

		return run(silent)("fossile", "open", fossilRepoName)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "fossil", "update")
	},
}

var BazaarBackend = &VCSBackend{
	// bazaar seems not supporting shallow clone currently.
	Clone: func(remote *url.URL, local string, ignoredShallow, silent bool) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return run(silent)("bzr", "branch", remote.String(), local)
	},
	Update: func(local string, silent bool) error {
		return runInDir(silent)(local, "bzr", "pull")
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
	"bazaar":     BazaarBackend,
}
