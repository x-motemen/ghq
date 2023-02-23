package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/x-motemen/ghq/cmdutil"
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
	url                              *url.URL
	dir                              string
	recursive, shallow, silent, bare bool
	branch                           string
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
		if vg.bare {
			args = append(args, "--bare")
		}
		args = append(args, vg.url.String(), vg.dir)

		return run(vg.silent)("git", args...)
	},
	Update: func(vg *vcsGetOption) error {
		if _, err := os.Stat(filepath.Join(vg.dir, ".git/svn")); err == nil {
			return GitsvnBackend.Update(vg)
		}
		err := runInDir(true)(vg.dir, "git", "rev-parse", "@{upstream}")
		if err != nil {
			err := runInDir(vg.silent)(vg.dir, "git", "fetch")
			if err != nil {
				return err
			}
			return nil
		}
		err = runInDir(vg.silent)(vg.dir, "git", "pull", "--ff-only")
		if err != nil {
			return err
		}
		if vg.recursive {
			return runInDir(vg.silent)(vg.dir, "git", "submodule", "update", "--init", "--recursive")
		}
		return nil
	},
	Init: func(dir string) error {
		return cmdutil.RunInDir(dir, "git", "init")
	},
	Contents: []string{".git"},
}

/*
If the svn target is under standard svn directory structure, "ghq" canonicalizes the checkout path.
For example, all following targets are checked-out into `$(ghq root)/svn.example.com/proj/repo`.

- svn.example.com/proj/repo
- svn.example.com/proj/repo/trunk
- svn.example.com/proj/repo/branches/featureN
- svn.example.com/proj/repo/tags/v1.0.1

Addition, when the svn target may be project root, "ghq" tries to checkout "/trunk".

This checkout rule is also applied when using "git-svn".
*/

const trunk = "/trunk"

var svnReg = regexp.MustCompile(`/(?:tags|branches)/[^/]+$`)

func replaceOnce(reg *regexp.Regexp, str, replace string) string {
	replaced := false
	return reg.ReplaceAllStringFunc(str, func(match string) string {
		if replaced {
			return match
		}
		replaced = true
		return reg.ReplaceAllString(match, replace)
	})
}

func svnBase(p string) string {
	if strings.HasSuffix(p, trunk) {
		return strings.TrimSuffix(p, trunk)
	}
	return replaceOnce(svnReg, p, "")
}

// SubversionBackend is the VCSBackend for subversion
var SubversionBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		vg.dir = svnBase(vg.dir)
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
			remote.Path = svnBase(remote.Path)
			remote.Path += "/branches/" + url.PathEscape(vg.branch)
		} else if !strings.HasSuffix(remote.Path, trunk) {
			copied := *vg.url
			copied.Path += trunk
			if err := cmdutil.RunSilently("svn", "info", copied.String()); err == nil {
				remote = &copied
			}
		}
		args = append(args, remote.String(), vg.dir)

		return run(vg.silent)("svn", args...)
	},
	Update: func(vg *vcsGetOption) error {
		return runInDir(vg.silent)(vg.dir, "svn", "update")
	},
	Contents: []string{".svn"},
}

var svnLastRevReg = regexp.MustCompile(`(?m)^Last Changed Rev: (\d+)$`)

// GitsvnBackend is the VCSBackend for git-svn
var GitsvnBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		orig := vg.dir
		vg.dir = svnBase(vg.dir)
		standard := orig == vg.dir

		dir, _ := filepath.Split(vg.dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		var getSvnInfo = func(u string) (string, error) {
			buf := &bytes.Buffer{}
			cmd := exec.Command("svn", "info", u)
			cmd.Stdout = buf
			cmd.Stderr = io.Discard
			err := cmdutil.RunCommand(cmd, true)
			return buf.String(), err
		}
		var svnInfo string
		args := []string{"svn", "clone"}
		remote := vg.url
		if vg.branch != "" {
			copied := *remote
			remote = &copied
			remote.Path = svnBase(remote.Path)
			remote.Path += "/branches/" + url.PathEscape(vg.branch)
			standard = false
		} else if standard {
			copied := *remote
			copied.Path += trunk
			info, err := getSvnInfo(copied.String())
			if err == nil {
				args = append(args, "-s")
				svnInfo = info
			} else {
				standard = false
			}
		}

		if vg.shallow {
			if svnInfo == "" {
				info, err := getSvnInfo(remote.String())
				if err != nil {
					return err
				}
				svnInfo = info
			}
			m := svnLastRevReg.FindStringSubmatch(svnInfo)
			if len(m) < 2 {
				return fmt.Errorf("no revisions are taken from svn info output: %s", svnInfo)
			}
			args = append(args, fmt.Sprintf("-r%s:HEAD", m[1]))
		}
		args = append(args, remote.String(), vg.dir)
		return run(vg.silent)("git", args...)
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
		return cmdutil.RunInDir(dir, "hg", "init")
	},
	Contents: []string{".hg"},
}

// DarcsBackend is the VCSBackend for darcs
var DarcsBackend = &VCSBackend{
	Clone: func(vg *vcsGetOption) error {
		if vg.branch != "" {
			return errors.New("darcs does not support branch")
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
		return cmdutil.RunInDir(dir, "darcs", "init")
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
			return errors.New("fossil does not support cloning specific branch")
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
		if err := cmdutil.RunInDir(dir, "fossil", "init", fossilRepoName); err != nil {
			return err
		}
		return cmdutil.RunInDir(dir, "fossil", "open", fossilRepoName)
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
		return cmdutil.RunInDir(dir, "bzr", "init")
	},
	Contents: []string{".bzr"},
}

var vcsRegistry = map[string]*VCSBackend{
	"git":        GitBackend,
	"github":     GitBackend,
	"codecommit": GitBackend,
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
