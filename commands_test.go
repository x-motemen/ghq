package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/motemen/ghq/cmdutil"
	"github.com/urfave/cli"
)

func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

type _cloneArgs struct {
	remote  *url.URL
	local   string
	shallow bool
}

type _updateArgs struct {
	local string
}

func withFakeGitBackend(t *testing.T, block func(*testing.T, string, *_cloneArgs, *_updateArgs)) {
	tmpRoot := newTempDir(t)
	defer os.RemoveAll(tmpRoot)

	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	_localRepositoryRoots = []string{tmpRoot}

	var cloneArgs _cloneArgs
	var updateArgs _updateArgs

	var originalGitBackend = GitBackend
	tmpBackend := &VCSBackend{
		Clone: func(remote *url.URL, local string, shallow, silent bool) error {
			cloneArgs = _cloneArgs{
				remote:  remote,
				local:   filepath.FromSlash(local),
				shallow: shallow,
			}
			return nil
		},
		Update: func(local string, silent bool) error {
			updateArgs = _updateArgs{
				local: local,
			}
			return nil
		},
	}
	GitBackend = tmpBackend
	vcsContentsMap[".git"] = tmpBackend
	defer func() { GitBackend = originalGitBackend; vcsContentsMap[".git"] = originalGitBackend }()

	block(t, tmpRoot, &cloneArgs, &updateArgs)
}

func TestCommandGet(t *testing.T) {
	app := newApp()

	testCases := []struct {
		name     string
		scenario func(*testing.T, string, *_cloneArgs, *_updateArgs)
	}{
		{
			name: "simple",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "motemen/ghq-test-repo"})

				expect := "https://github.com/motemen/ghq-test-repo"
				if cloneArgs.remote.String() != expect {
					t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
				}
				if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
					t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
				}
				if cloneArgs.shallow {
					t.Errorf("cloneArgs.shallow should be false")
				}
			},
		},
		{
			name: "-p option",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "-p", "motemen/ghq-test-repo"})

				expect := "ssh://git@github.com/motemen/ghq-test-repo"
				if cloneArgs.remote.String() != expect {
					t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
				}
				if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
					t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
				}
				if cloneArgs.shallow {
					t.Errorf("cloneArgs.shallow should be false")
				}
			},
		},
		{
			name: "already cloned with -u",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
				// mark as "already cloned", the condition may change later
				os.MkdirAll(filepath.Join(localDir, ".git"), 0755)

				app.Run([]string{"", "get", "-u", "motemen/ghq-test-repo"})

				if updateArgs.local != localDir {
					t.Errorf("got: %s, expect: %s", updateArgs.local, localDir)
				}
			},
		},
		{
			name: "shallow",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "-shallow", "motemen/ghq-test-repo"})

				expect := "https://github.com/motemen/ghq-test-repo"
				if cloneArgs.remote.String() != expect {
					t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
				}
				if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
					t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
				}
				if !cloneArgs.shallow {
					t.Errorf("cloneArgs.shallow should be true")
				}
			},
		},
		{
			name: "dot slach ./",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen")
				os.MkdirAll(localDir, 0755)
				wd, _ := os.Getwd()
				os.Chdir(localDir)
				defer os.Chdir(wd)

				app.Run([]string{"", "get", "-u", "." + string(filepath.Separator) + "ghq-test-repo"})

				expect := "https://github.com/motemen/ghq-test-repo"
				if cloneArgs.remote.String() != expect {
					t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
				}
				expectDir := filepath.Join(localDir, "ghq-test-repo")
				if cloneArgs.local != expectDir {
					t.Errorf("got: %s, expect: %s", cloneArgs.local, expectDir)
				}
			},
		},
		{
			name: "dot dot slash ../",
			scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
				os.MkdirAll(localDir, 0755)
				wd, _ := os.Getwd()
				os.Chdir(localDir)
				defer os.Chdir(wd)

				app.Run([]string{"", "get", "-u", ".." + string(filepath.Separator) + "ghq-another-test-repo"})

				expect := "https://github.com/motemen/ghq-another-test-repo"
				if cloneArgs.remote.String() != expect {
					t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
				}
				expectDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-another-test-repo")
				if cloneArgs.local != expectDir {
					t.Errorf("got: %s, expect: %s", cloneArgs.local, expectDir)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withFakeGitBackend(t, tc.scenario)
		})
	}
}

func TestDoRoot(t *testing.T) {
	ghqrootEnv := "GHQ_ROOT"
	testCases := []struct {
		name              string
		setup             func() func()
		expect, allExpect string
	}{{
		name: "env",
		setup: func() func() {
			orig := os.Getenv(ghqrootEnv)
			os.Setenv(ghqrootEnv, "/path/to/ghqroot1"+string(os.PathListSeparator)+"/path/to/ghqroot2")
			return func() { os.Setenv(ghqrootEnv, orig) }
		},
		expect:    "/path/to/ghqroot1\n",
		allExpect: "/path/to/ghqroot1\n/path/to/ghqroot2\n",
	}, {
		name: "gitconfig",
		setup: func() func() {
			orig := os.Getenv(ghqrootEnv)
			os.Setenv(ghqrootEnv, "")
			teardown := withGitConfig(t, `[ghq]
  root = /path/to/ghqroot11
  root = /path/to/ghqroot12
`)
			return func() {
				os.Setenv(ghqrootEnv, orig)
				teardown()
			}
		},
		expect:    "/path/to/ghqroot11\n",
		allExpect: "/path/to/ghqroot11\n/path/to/ghqroot12\n",
	}, {
		name: "default home",
		setup: func() func() {
			origRoot := os.Getenv(ghqrootEnv)
			os.Setenv(ghqrootEnv, "")
			origGitconfig := os.Getenv("GIT_CONFIG")
			os.Setenv("GIT_CONFIG", "/tmp/unknown-ghq-dummy")
			origHome := os.Getenv("HOME")
			os.Setenv("HOME", "/path/to/ghqhome")

			return func() {
				os.Setenv(ghqrootEnv, origRoot)
				os.Setenv("GIT_CONFIG", origGitconfig)
				os.Setenv("HOME", origHome)
			}
		},
		expect:    "/path/to/ghqhome/.ghq\n",
		allExpect: "/path/to/ghqhome/.ghq\n",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
			_localRepositoryRoots = nil
			defer tc.setup()()
			out, _, _ := capture(func() {
				newApp().Run([]string{"", "root"})
			})
			if out != tc.expect {
				t.Errorf("got: %s, expect: %s", out, tc.expect)
			}
			out, _, _ = capture(func() {
				newApp().Run([]string{"", "root", "--all"})
			})
			if out != tc.allExpect {
				t.Errorf("got: %s, expect: %s", out, tc.allExpect)
			}
		})
	}
}

func TestDoLook(t *testing.T) {
	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "ghq", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "gobump", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "Songmu", "gobump", ".git"), 0755)
		defer func(orig func(cmd *exec.Cmd) error) {
			cmdutil.CommandRunner = orig
		}(cmdutil.CommandRunner)
		var lastCmd *exec.Cmd
		cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
			lastCmd = cmd
			return nil
		}
		sh := detectShell()

		err := newApp().Run([]string{"", "look", "https://github.com/motemen/ghq"})
		if err != nil {
			t.Errorf("error should be nil, but: %s", err)
		}

		if !reflect.DeepEqual(lastCmd.Args, []string{sh}) {
			t.Errorf("lastCmd.Args: got: %v, expect: %v", lastCmd.Args, []string{sh})
		}
		dir := filepath.Join(tmproot, "github.com", "motemen", "ghq")
		if lastCmd.Dir != dir {
			t.Errorf("lastCmd.Dir: got: %s, expect: %s", lastCmd.Dir, dir)
		}
		gotEnv := lastCmd.Env[len(lastCmd.Env)-1]
		expectEnv := "GHQ_LOOK=github.com/motemen/ghq"
		if gotEnv != expectEnv {
			t.Errorf("lastCmd.Env[len(lastCmd.Env)-1]: got: %s, expect: %s", gotEnv, expectEnv)
		}

		err = newApp().Run([]string{"", "look", "github.com/motemen/_unknown"})
		expect := "No repository found"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}

		err = newApp().Run([]string{"", "look", "gobump"})
		expect = "More than one repositories are found; Try more precise name"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}
	})
}
