package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Songmu/gitconfig"
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
			defer os.RemoveAll(tmpd)
			defer gitconfig.WithConfig(t, fmt.Sprintf(`
[ghq "https://github.com/motemen"]
  root = "%s"
`, filepath.ToSlash(tmpd)))()
			app.Run([]string{"", "get", "motemen/ghq-test-repo"})

			localDir := filepath.Join(tmpd, "github.com", "motemen", "ghq-test-repo")
			if filepath.ToSlash(cloneArgs.local) != filepath.ToSlash(localDir) {
				t.Errorf("got: %s, expect: %s", filepath.ToSlash(cloneArgs.local), filepath.ToSlash(localDir))
			}
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withFakeGitBackend(t, tc.scenario)
		})
	}
}
