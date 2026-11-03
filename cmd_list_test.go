package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/x-motemen/ghq/logger"
)

func TestCommandList(t *testing.T) {
	_, _, err := capture(func() {
		newApp().Run(context.Background(), []string{"ghq", "list"})
	})

	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}

func TestCommandListUnique(t *testing.T) {
	_, _, err := capture(func() {
		newApp().Run(context.Background(), []string{"ghq", "list", "--unique"})
	})

	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}

func TestCommandListUnknown(t *testing.T) {
	_, _, err := capture(func() {
		newApp().Run(context.Background(), []string{"ghq", "list", "--unknown-flag"})
	})

	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}

func sortLines(s string) string {
	ss := strings.Split(strings.TrimSpace(s), "\n")
	sort.Strings(ss)
	return strings.Join(ss, "\n")
}

func equalPathLines(lhs, rhs string) bool {
	return sortLines(lhs) == sortLines(rhs)
}

func TestDoList_query(t *testing.T) {
	gitRepos := []string{
		"github.com/motemen/ghq",
		"github.com/motemen/gobump",
		"github.com/motemen/gore",
		"github.com/Songmu/gobump",
		"github.com/test/Awesome",
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
		args:   []string{"-exact", "gobump"},
		expect: "github.com/Songmu/gobump\ngithub.com/motemen/gobump\n",
	}, {
		name:   "query",
		args:   []string{"men/go"},
		expect: "github.com/motemen/gobump\ngithub.com/motemen/gore\n",
	}, {
		name:   "exact query",
		args:   []string{"-exact", "men/go"},
		expect: "",
	}, {
		name:   "vcs",
		args:   []string{"--vcs", "svn"},
		expect: "github.com/msh5/svntest\n",
	}, {
		name:   "smartcasing fuzzy",
		args:   []string{"awesome"},
		expect: "github.com/test/Awesome\n",
	}, {
		name:   "smartcasing exact",
		args:   []string{"Awesome"},
		expect: "github.com/test/Awesome\n",
	}, {
		name:   "smartcasing exact fail",
		args:   []string{"aWesome"},
		expect: "",
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
					newApp().Run(context.Background(), args)
				})
				if !equalPathLines(out, tc.expect) {
					t.Errorf("got:\n%s\nexpect:\n%s", out, tc.expect)
				}
				if strings.Contains(tc.name, "unique") {
					return
				}
				argsFull := append([]string{"ghq", "list", "--full-path"}, tc.args...)
				fullExpect := tc.expect
				if fullExpect != "" {
					if runtime.GOOS == "windows" {
						fullExpect = strings.ReplaceAll(fullExpect, `/`, `\`)
					}
					fullExpect = tmproot + string(filepath.Separator) + strings.TrimSpace(fullExpect)
					fullExpect = strings.ReplaceAll(fullExpect, "\n", "\n"+tmproot+string(filepath.Separator))
					fullExpect += "\n"
				}
				out, _, _ = capture(func() {
					newApp().Run(context.Background(), argsFull)
				})
				if !equalPathLines(out, fullExpect) {
					t.Errorf("got:\n%s\nexpect:\n%s", out, fullExpect)
				}
			})
		}
	})
}

func TestDoList_unique(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)

	tmp1 := newTempDir(t)
	tmp2 := newTempDir(t)

	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}
	rootPaths := []string{tmp1, tmp2}
	setEnv(t, envGhqRoot, strings.Join(rootPaths, string(os.PathListSeparator)))
	for _, rootPath := range rootPaths {
		os.MkdirAll(filepath.Join(rootPath, "github.com/motemen/ghq/.git"), 0755)
	}
	out, _, _ := capture(func() {
		newApp().Run(context.Background(), []string{"ghq", "list", "--unique"})
	})
	if out != "ghq\n" {
		t.Errorf("got: %s, expect: ghq\n", out)
	}
}

func TestDoList_unknownRoot(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, "/path/to/unknown-ghq")
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	err := newApp().Run(context.Background(), []string{"ghq", "list"})
	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}

func TestDoList_notPermittedRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	tmpdir := newTempDir(t)
	defer os.Chmod(tmpdir, 0755)
	setEnv(t, envGhqRoot, tmpdir)

	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}
	os.Chmod(tmpdir, 0000)

	err := newApp().Run(context.Background(), []string{"ghq", "list"})
	if err != nil {
		t.Errorf("error should be nil, but: %+v", err)
	}
}

func TestDoList_unsearchableRootParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires Unix directory permissions")
	}
	parent := newTempDir(t)
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	otherRoot := newTempDir(t)
	if err := os.MkdirAll(filepath.Join(otherRoot, "github.com/motemen/ghq/.git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(parent, 0755)
	if _, err := os.Stat(root); !os.IsPermission(err) {
		t.Skipf("cannot reproduce permission-denied stat: %v", err)
	}
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	t.Setenv(envGhqRoot, strings.Join([]string{root, otherRoot}, string(os.PathListSeparator)))
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	var out, logs bytes.Buffer
	app := newApp()
	app.Writer = &out
	logger.SetOutput(&logs)
	defer logger.SetOutput(os.Stderr)
	if err := app.Run(context.Background(), []string{"ghq", "list"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "github.com/motemen/ghq\n" {
		t.Errorf("got: %q, expect repository from accessible root", out.String())
	}
	if !strings.Contains(logs.String(), root+": Permission denied") {
		t.Errorf("missing permission warning for %s: %s", root, &logs)
	}
}

func TestDoList_rootStatError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires symlinks")
	}
	root := filepath.Join(newTempDir(t), "loop")
	if err := os.Symlink(root, root); err != nil {
		t.Fatal(err)
	}
	_, statErr := os.Stat(root)
	if statErr == nil || os.IsNotExist(statErr) || os.IsPermission(statErr) {
		t.Fatalf("expected a non-permission stat error, got: %v", statErr)
	}
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	t.Setenv(envGhqRoot, root)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	err := newApp().Run(context.Background(), []string{"ghq", "list"})
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("expected stat error, got: %v", err)
	}
	if pathErr.Op != "stat" || pathErr.Path != root || !errors.Is(err, errors.Unwrap(statErr)) {
		t.Errorf("expected original stat error for %s, got: %v", root, err)
	}
}

func TestDoList_withSystemHiddenDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	tmpdir := newTempDir(t)
	systemHidden := filepath.Join(tmpdir, ".system")
	os.MkdirAll(systemHidden, 0000)
	defer os.Chmod(systemHidden, 0755)
	setEnv(t, envGhqRoot, tmpdir)

	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	err := newApp().Run(context.Background(), []string{"ghq", "list"})
	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}
}
