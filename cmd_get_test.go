package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Songmu/gitconfig"
	"github.com/x-motemen/ghq/cmdutil"
	"github.com/x-motemen/ghq/logger"
)

func TestCommandGet(t *testing.T) {
	app := newApp()

	testCases := []struct {
		name     string
		scenario func(*testing.T, string, *_cloneArgs, *_updateArgs)
	}{{
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
			if cloneArgs.branch != "" {
				t.Errorf("cloneArgs.branch should be empty")
			}
			if !cloneArgs.recursive {
				t.Errorf("cloneArgs.recursive should be true")
			}
			if cloneArgs.bare {
				t.Errorf("cloneArgs.bare should be false")
			}
		},
	}, {
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
	}, {
		name: "already cloned with -update",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
			// mark as "already cloned", the condition may change later
			os.MkdirAll(filepath.Join(localDir, ".git"), 0755)

			app.Run([]string{"", "get", "-update", "motemen/ghq-test-repo"})

			if updateArgs.local != localDir {
				t.Errorf("got: %s, expect: %s", updateArgs.local, localDir)
			}
		},
	}, {
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
	}, {
		name: "dot slash ./",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen")
			os.MkdirAll(localDir, 0755)
			wd, _ := os.Getwd()
			os.Chdir(localDir)
			defer os.Chdir(wd)

			app.Run([]string{"", "get", "-update", "." + string(filepath.Separator) + "ghq-test-repo"})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			expectDir := filepath.Join(localDir, "ghq-test-repo")
			if cloneArgs.local != expectDir {
				t.Errorf("got: %s, expect: %s", cloneArgs.local, expectDir)
			}
		},
	}, {
		name: "dot dot slash ../",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
			os.MkdirAll(localDir, 0755)
			wd, _ := os.Getwd()
			os.Chdir(localDir)
			defer os.Chdir(wd)

			app.Run([]string{"", "get", "-update", ".." + string(filepath.Separator) + "ghq-another-test-repo"})

			expect := "https://github.com/motemen/ghq-another-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			expectDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-another-test-repo")
			if cloneArgs.local != expectDir {
				t.Errorf("got: %s, expect: %s", cloneArgs.local, expectDir)
			}
		},
	}, {
		name: "specific branch",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

			expectBranch := "hello"
			app.Run([]string{"", "get", "-shallow", "-branch", expectBranch, "motemen/ghq-test-repo"})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if cloneArgs.branch != expectBranch {
				t.Errorf("got: %q, expect: %q", cloneArgs.branch, expectBranch)
			}
		},
	}, {
		name: "specific branch using @ syntax",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

			expectBranch := "hello"
			app.Run([]string{"", "get", "-shallow", "motemen/ghq-test-repo@" + expectBranch})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if cloneArgs.branch != expectBranch {
				t.Errorf("got: %q, expect: %q", cloneArgs.branch, expectBranch)
			}
		},
	}, {
		name: "with --no-recursive option",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			app.Run([]string{"", "get", "--no-recursive", "motemen/ghq-test-repo"})

			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if cloneArgs.recursive {
				t.Errorf("cloneArgs.recursive should be false")
			}
		},
	}, {
		name: "ghq.<url>.root",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			tmpd := newTempDir(t)
			t.Cleanup(gitconfig.WithConfig(t, fmt.Sprintf(`
[ghq "https://github.com/motemen"]
root = "%s"
`, filepath.ToSlash(tmpd))))
			app.Run([]string{"", "get", "motemen/ghq-test-repo"})

			localDir := filepath.Join(tmpd, "github.com", "motemen", "ghq-test-repo")
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
		},
	}, {
		name: "ghq<url>.hostFolderName",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			t.Cleanup(gitconfig.WithConfig(t, `
[ghq "https://github.com"]
	 hostFolderName = gh
`))
			app.Run([]string{"", "get", "motemen/ghq-test-repo"})

			localDir := filepath.Join(tmpRoot, "gh", "motemen", "ghq-test-repo")
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
		},
	}, {
		name: "bare",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo.git")

			app.Run([]string{"", "get", "--bare", "motemen/ghq-test-repo"})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if !cloneArgs.bare {
				t.Errorf("cloneArgs.bare should be true")
			}
		},
	}, {
		name: "silent mode",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

			out, _, err := captureWithInput([]string{}, func() {
				app.Run([]string{"", "get", "--silent", "motemen/ghq-test-repo"})
			})
			if err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}

			if !cloneArgs.silent {
				t.Errorf("cloneArgs.silent should be true")
			}
			if out != "" {
				t.Errorf("silent mode should not output any logs, but got: %s", out)
			}
		},
	}, {
		name: "[partial] blobless",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

			app.Run([]string{"", "get", "--partial", "blobless", "motemen/ghq-test-repo"})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if cloneArgs.partial != "blobless" {
				t.Errorf("cloneArgs.partial should be \"blobless\"")
			}
		},
	}, {
		name: "[partial] treeless",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

			app.Run([]string{"", "get", "--partial", "treeless", "motemen/ghq-test-repo"})

			expect := "https://github.com/motemen/ghq-test-repo"
			if cloneArgs.remote.String() != expect {
				t.Errorf("got: %s, expect: %s", cloneArgs.remote, expect)
			}
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
			if cloneArgs.partial != "treeless" {
				t.Errorf("cloneArgs.partial should be \"treeless\"")
			}
		},
	}, {
		name: "[partial] unacceptable value",
		scenario: func(t *testing.T, tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
			err := app.Run([]string{"", "get", "--partial", "unacceptable", "motemen/ghq-test-repo"})

			expect := "flag partial value \"unacceptable\" is not allowed"
			if err.Error() != expect {
				t.Errorf("got: %s, expect: %s", err.Error(), expect)
			}
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withFakeGitBackend(t, tc.scenario)
		})
	}
}

func TestLook(t *testing.T) {
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

		err := newApp().Run([]string{"", "get", "--look", "https://github.com/motemen/ghq"})
		if err != nil {
			t.Errorf("error should be nil, but: %s", err)
		}

		if !reflect.DeepEqual(lastCmd.Args, []string{sh}) {
			t.Errorf("lastCmd.Args: got: %v, expect: %v", lastCmd.Args, []string{sh})
		}
		dir := filepath.Join(tmproot, "github.com", "motemen", "ghq")
		if filepath.Clean(lastCmd.Dir) != dir {
			t.Errorf("lastCmd.Dir: got: %s, expect: %s", lastCmd.Dir, dir)
		}
		gotEnv := lastCmd.Env[len(lastCmd.Env)-1]
		expectEnv := "GHQ_LOOK=github.com/motemen/ghq"
		if gotEnv != expectEnv {
			t.Errorf("lastCmd.Env[len(lastCmd.Env)-1]: got: %s, expect: %s", gotEnv, expectEnv)
		}

		err = look("github.com/motemen/_unknown", false)
		expect := "no repository found"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}

		err = look("gobump", false)
		expect = "More than one repositories are found; Try more precise name"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}
	})
}

func TestBareLook(t *testing.T) {
	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "ghq.git"), 0o755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "gobump", ".git"), 0o755)
		defer func(orig func(cmd *exec.Cmd) error) {
			cmdutil.CommandRunner = orig
		}(cmdutil.CommandRunner)
		var lastCmd *exec.Cmd
		cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
			lastCmd = cmd
			return nil
		}
		sh := detectShell()

		err := newApp().Run([]string{"", "get", "--bare", "--look", "https://github.com/motemen/ghq.git"})
		if err != nil {
			t.Errorf("error should be nil, but: %s", err)
		}

		if !reflect.DeepEqual(lastCmd.Args, []string{sh}) {
			t.Errorf("lastCmd.Args: got: %v, expect: %v", lastCmd.Args, []string{sh})
		}
		dir := filepath.Join(tmproot, "github.com", "motemen", "ghq.git")
		if filepath.Clean(lastCmd.Dir) != dir {
			t.Errorf("lastCmd.Dir: got: %s, expect: %s", lastCmd.Dir, dir)
		}
		gotEnv := lastCmd.Env[len(lastCmd.Env)-1]
		expectEnv := "GHQ_LOOK=github.com/motemen/ghq.git"
		if gotEnv != expectEnv {
			t.Errorf("lastCmd.Env[len(lastCmd.Env)-1]: got: %s, expect: %s", gotEnv, expectEnv)
		}

		err = look("github.com/motemen/ghq", false)
		expect := "no repository found"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}

		err = look("github.com/motemen/gobump.git", true)
		expect = "no repository found"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}
	})
}

func TestDoGet_bulk(t *testing.T) {
	in := []string{
		"github.com/x-motemen/ghq",
		"github.com/motemen/gore",
	}

	testCases := []struct {
		name string
		args []string
	}{{
		name: "normal",
		args: []string{},
	}, {
		name: "parallel",
		args: []string{"-parallel"},
	}}

	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	defer func() { logger.SetOutput(os.Stderr) }()

	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		for _, r := range in {
			os.MkdirAll(filepath.Join(tmproot, r, ".git"), 0755)
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf.Reset()
				out, _, err := captureWithInput(in, func() {
					args := append([]string{"", "get"}, tc.args...)
					if err := newApp().Run(args); err != nil {
						t.Errorf("error should be nil but: %s", err)
					}
				})
				if err != nil {
					t.Errorf("error should be nil, but: %s", err)
				}
				if out != "" {
					t.Errorf("out should be empty, but: %s", out)
				}
				log := filepath.ToSlash(buf.String())
				for _, r := range in {
					if !strings.Contains(log, r) {
						t.Errorf("log should contains %q but not: %s", r, log)
					}
				}
			})
		}
	})
}
