package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

func TestCommandList(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
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
		t.Errorf("error should be nil, but: %v", err)
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
		t.Errorf("error should be nil, but: %v", err)
	}
}

func sortLines(s string) string {
	ss := strings.Split(s, "\n")
	sort.Strings(ss)
	return strings.Join(ss, "\n")
}

func TestDoList_query(t *testing.T) {
	gitRepos := []string{
		"github.com/motemen/ghq",
		"github.com/motemen/gobump",
		"github.com/motemen/gore",
		"github.com/Songmu/gobump",
		"golang.org/x/crypt",
		"golang.org/x/image",
	}
	svnRepos := []string{
		"github.com/msh5/svntest",
	}
	testCases := []struct {
		name   string
		args   []string
		expect string
	}{{
		name:   "repo match",
		args:   []string{"ghq"},
		expect: "github.com/motemen/ghq\n",
	}, {
		name:   "unique",
		args:   []string{"--unique", "ghq"},
		expect: "ghq\n",
	}, {
		name:   "host only doesn't match",
		args:   []string{"github.com"},
		expect: "",
	}, {
		name:   "host and slash match",
		args:   []string{"golang.org/"},
		expect: "golang.org/x/crypt\ngolang.org/x/image\n",
	}, {
		name:   "host and user",
		args:   []string{"github.com/Songmu"},
		expect: "github.com/Songmu/gobump\n",
	}, {
		name:   "with scheme",
		args:   []string{"https://github.com/motemen/ghq"},
		expect: "github.com/motemen/ghq\n",
	}, {
		name:   "exact",
		args:   []string{"-e", "gobump"},
		expect: "github.com/Songmu/gobump\ngithub.com/motemen/gobump\n",
	}, {
		name:   "query",
		args:   []string{"men/go"},
		expect: "github.com/motemen/gobump\ngithub.com/motemen/gore\n",
	}, {
		name:   "exact query",
		args:   []string{"-e", "men/go"},
		expect: "",
	}, {
		name:   "vcs",
		args:   []string{"--vcs", "svn"},
		expect: "github.com/msh5/svntest\n",
	}}

	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		for _, r := range gitRepos {
			os.MkdirAll(filepath.Join(tmproot, r, ".git"), 0755)
		}
		for _, r := range svnRepos {
			os.MkdirAll(filepath.Join(tmproot, r, ".svn"), 0755)
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				args := append([]string{"ghq", "list"}, tc.args...)
				out, _, _ := capture(func() {
					newApp().Run(args)
				})
				if sortLines(out) != sortLines(tc.expect) {
					t.Errorf("got:\n%s\nexpect:\n%s", out, tc.expect)
				}
				if strings.Contains(tc.name, "unique") {
					return
				}
				argsFull := append([]string{"ghq", "list", "--full-path"}, tc.args...)
				fullExpect := tc.expect
				if fullExpect != "" {
					fullExpect = tmproot + string(filepath.Separator) + strings.TrimSpace(fullExpect)
					fullExpect = strings.ReplaceAll(fullExpect, "\n", "\n"+tmproot+string(filepath.Separator))
					fullExpect += "\n"
				}
				out, _, _ = capture(func() {
					newApp().Run(argsFull)
				})
				if sortLines(out) != sortLines(fullExpect) {
					t.Errorf("got:\n%s\nexpect:\n%s", out, fullExpect)
				}
			})
		}
	})
}

func TestDoList_unique(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	defer func(orig string) { os.Setenv("GHQ_ROOT", orig) }(os.Getenv("GHQ_ROOT"))

	tmp1 := newTempDir(t)
	defer os.RemoveAll(tmp1)
	tmp2 := newTempDir(t)
	defer os.RemoveAll(tmp2)

	_localRepositoryRoots = nil
	rootPaths := []string{tmp1, tmp2}
	os.Setenv("GHQ_ROOT", strings.Join(rootPaths, string(os.PathListSeparator)))
	for _, rootPath := range rootPaths {
		os.MkdirAll(filepath.Join(rootPath, "github.com/motemen/ghq/.git"), 0755)
	}
	out, _, _ := capture(func() {
		newApp().Run([]string{"ghq", "list", "--unique"})
	})
	if out != "ghq\n" {
		t.Errorf("got: %s, expect: ghq\n", out)
	}
}

func TestDoList_unknownRoot(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	defer tmpEnv("GHQ_ROOT", "/path/to/unknown-ghq")()
	_localRepositoryRoots = nil

	err := newApp().Run([]string{"ghq", "list"})
	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}

func TestDoList_notPermittedRoot(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	tmpdir := newTempDir(t)
	defer func(dir string) {
		os.Chmod(dir, 0755)
		os.RemoveAll(dir)
	}(tmpdir)
	defer tmpEnv("GHQ_ROOT", tmpdir)()

	_localRepositoryRoots = nil
	os.Chmod(tmpdir, 0000)

	err := newApp().Run([]string{"ghq", "list"})
	if !os.IsPermission(xerrors.Unwrap(err)) {
		t.Errorf("error should be ErrPermission, but: %#+v", err)
	}
}

func TestDoList_withSystemHiddenDir(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	tmpdir := newTempDir(t)
	systemHidden := filepath.Join(tmpdir, ".system")
	os.MkdirAll(systemHidden, 0000)
	defer func(dir string) {
		os.Chmod(systemHidden, 0755)
		os.RemoveAll(dir)
	}(tmpdir)
	defer tmpEnv("GHQ_ROOT", tmpdir)()

	_localRepositoryRoots = nil

	err := newApp().Run([]string{"ghq", "list"})
	if err == nil {
		t.Errorf("should not be error")
	}
}
