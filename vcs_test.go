package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/x-motemen/ghq/cmdutil"
)

var (
	remoteDummyURL = mustParseURL("https://example.com/git/repo")
	dummySvnInfo   = []byte(`Path: trunk
URL: https://svn.apache.org/repos/asf/subversion/trunk
Relative URL: ^/subversion/trunk
Repository Root: https://svn.apache.org/repos/asf
Repository UUID: 13f79535-47bb-0310-9956-ffa450edef68
Revision: 1872085
Node Kind: directory
Last Changed Author: julianfoad
Last Changed Rev: 1872031
Last Changed Date: 2019-08-16 15:16:45 +0900 (Fri, 16 Aug 2019)
`)
)

func TestVCSBackend(t *testing.T) {
	tempDir := newTempDir(t)
	defer os.RemoveAll(tempDir)
	localDir := filepath.Join(tempDir, "repo")
	_commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return _commands[len(_commands)-1] }
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		_commands = append(_commands, cmd)
		if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
			return fmt.Errorf("[test] failed to svn info")
		}
		return nil
	}

	testCases := []struct {
		name   string
		f      func() error
		expect []string
		dir    string
	}{{
		name: "[git] clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"git", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] shallow clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
				silent:  true,
			})
		},
		expect: []string{"git", "clone", "--depth", "1", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] clone specific branch",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"git", "clone", "--branch", "hello", "--single-branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] update",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				return nil
			}
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "pull", "--ff-only"},
		dir:    localDir,
	}, {
		name: "[git] fetch",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"git", "rev-parse", "@{upstream}"}) {
					return fmt.Errorf("[test] failed to git rev-parse @{upstream}")
				}
				return nil
			}
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "fetch"},
		dir:    localDir,
	}, {
		name: "[git] recursive",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:       remoteDummyURL,
				dir:       localDir,
				recursive: true,
			})
		},
		expect: []string{"git", "clone", "--recursive", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] update recursive",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				return nil
			}
			return GitBackend.Update(&vcsGetOption{
				dir:       localDir,
				recursive: true,
			})
		},
		expect: []string{"git", "submodule", "update", "--init", "--recursive"},
		dir:    localDir,
	}, {
		name: "[git] switch git-svn on update",
		f: func() error {
			err := os.MkdirAll(filepath.Join(localDir, ".git", "svn"), 0755)
			if err != nil {
				return err
			}
			defer os.RemoveAll(filepath.Join(localDir, ".git"))
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "rebase"},
		dir:    localDir,
	}, {
		name: "[svn] checkout",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] checkout shallow",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"svn", "checkout", "--depth", "immediates", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] checkout specific branch",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String() + "/branches/hello", localDir},
	}, {
		name: "[svn] checkout with filling trunk",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			return SubversionBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String() + "/trunk", localDir},
	}, {
		name: "[svn] update",
		f: func() error {
			return SubversionBackend.Update(&vcsGetOption{
				dir:    localDir,
				silent: true,
			})
		},
		expect: []string{"svn", "update"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone",
		f: func() error {
			return GitsvnBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git-svn] update",
		f: func() error {
			return GitsvnBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "rebase"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone shallow",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			return GitsvnBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"git", "svn", "clone", "-s", "-r1872031:HEAD", remoteDummyURL.String(), localDir},
	}, {
		name: "[git-svn] clone specific branch",
		f: func() error {
			return GitsvnBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String() + "/branches/hello", localDir},
	}, {
		name: "[git-svn] clone specific branch from tagged URL with shallow",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(
					cmd.Args, []string{"svn", "info", "https://example.com/git/repo/branches/develop"},
				) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			copied := *remoteDummyURL
			copied.Path += "/tags/v9.9.9"
			return GitsvnBackend.Clone(&vcsGetOption{
				url:     &copied,
				dir:     localDir,
				branch:  "develop",
				shallow: true,
			})
		},
		expect: []string{
			"git", "svn", "clone", "-r1872031:HEAD", remoteDummyURL.String() + "/branches/develop", localDir},
	}, {
		name: "[hg] clone",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] update",
		f: func() error {
			return MercurialBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"hg", "pull", "--update"},
		dir:    localDir,
	}, {
		name: "[hg] clone shallow",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] clone specific branch",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"hg", "clone", "--branch", "hello", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone",
		f: func() error {
			return DarcsBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"darcs", "get", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone shallow",
		f: func() error {
			return DarcsBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"darcs", "get", "--lazy", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] update",
		f: func() error {
			return DarcsBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"darcs", "pull"},
		dir:    localDir,
	}, {
		name: "[bzr] clone",
		f: func() error {
			return BazaarBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[bzr] update",
		f: func() error {
			return BazaarBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"bzr", "pull", "--overwrite"},
		dir:    localDir,
	}, {
		name: "[bzr] clone shallow",
		f: func() error {
			return BazaarBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[fossil] clone",
		f: func() error {
			return FossilBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"fossil", "open", fossilRepoName},
		dir:    localDir,
	}, {
		name: "[fossil] update",
		f: func() error {
			return FossilBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"fossil", "update"},
		dir:    localDir,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.f(); err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}
			c := lastCommand()
			if !reflect.DeepEqual(c.Args, tc.expect) {
				t.Errorf("\ngot:    %+v\nexpect: %+v", c.Args, tc.expect)
			}
			if c.Dir != tc.dir {
				t.Errorf("got: %s, expect: %s", c.Dir, tc.dir)
			}
		})
	}
}

func TestCvsDummyBackend(t *testing.T) {
	tempDir := newTempDir(t)
	defer os.RemoveAll(tempDir)
	localDir := filepath.Join(tempDir, "repo")

	if err := cvsDummyBackend.Clone(&vcsGetOption{
		url: remoteDummyURL,
		dir: localDir,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := cvsDummyBackend.Clone(&vcsGetOption{
		url:     remoteDummyURL,
		dir:     localDir,
		shallow: true,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := cvsDummyBackend.Update(&vcsGetOption{
		dir: localDir,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}
}

func TestBranchOptionIgnoredErrors(t *testing.T) {
	tempDir := newTempDir(t)
	defer os.RemoveAll(tempDir)
	localDir := filepath.Join(tempDir, "repo")

	if err := DarcsBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := FossilBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := BazaarBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}
}
