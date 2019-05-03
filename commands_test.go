package main

import (
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli"
)

func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func captureReader(block func()) (*os.File, *os.File, error) {
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	rErr, wErr, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	defer wOut.Close()
	defer wErr.Close()

	var stdout, stderr *os.File
	os.Stdout, stdout = wOut, os.Stdout
	os.Stderr, stderr = wErr, os.Stderr

	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	block()

	wOut.Close()
	wErr.Close()

	return rOut, rErr, nil
}

func capture(block func()) (string, string, error) {
	rOut, rErr, err := captureReader(block)
	if err != nil {
		return "", "", err
	}
	defer rOut.Close()
	defer rErr.Close()

	bufOut, err := ioutil.ReadAll(rOut)
	if err != nil {
		return "", "", err
	}

	bufErr, err := ioutil.ReadAll(rErr)
	if err != nil {
		return "", "", err
	}

	return string(bufOut), string(bufErr), nil
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
	tmpRoot, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatalf("Could not create tempdir: %s", err)
	}
	defer os.RemoveAll(tmpRoot)

	// Resolve /var/folders/.../T/... to /private/var/... in OSX
	tmpRoot = func() string {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd(): %s", err)
		}

		defer func() {
			err := os.Chdir(wd)
			if err != nil {
				t.Fatalf("os.Chdir(): %s", err)
			}
		}()

		err = os.Chdir(tmpRoot)
		if err != nil {
			t.Fatalf("os.Chdir(): %s", err)
		}

		tmpRoot, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd(): %s", err)
		}

		return tmpRoot
	}()

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
	vcsDirsMap[".git"] = tmpBackend
	defer func() { GitBackend = originalGitBackend; vcsDirsMap[".git"] = originalGitBackend }()

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

func TestCommandList(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
	}
}

func TestCommandListUnique(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unique"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
	}
}

func TestCommandListUnknown(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unknown-flag"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
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
			teardown, err := WithGitconfigFile(`[ghq]
  root = /path/to/ghqroot11
  root = /path/to/ghqroot12
`)
			if err != nil {
				panic(err)
			}
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
