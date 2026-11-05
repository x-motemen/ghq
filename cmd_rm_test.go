package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Songmu/gitconfig"
	"github.com/x-motemen/ghq/cmdutil"
)

func TestRmCommand(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	commandRunner := func(cmd *exec.Cmd) error {
		return nil
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpd := newTempDir(t)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpd)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	testCases := []struct {
		name      string
		input     []string
		setup     func(t *testing.T)
		expectErr bool
		cmdRun    func(cmd *exec.Cmd) error
		skipOnWin bool
	}{
		{
			name:  "simple",
			input: []string{"rm", "motemen/ghqq"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: false,
		},
		{
			name:  "unknown host with one path component",
			input: []string{"rm", "https://example.com/repo"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "example.com", "repo", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:  "empty directory",
			input: []string{"rm", "motemen/ghqqq"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqqq"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: true,
		},
		{
			name:      "missing directory",
			input:     []string{"rm", "motemen/missing"},
			setup:     func(t *testing.T) {},
			expectErr: true,
		},
		{
			name:  "incorrect repository name",
			input: []string{"rm", "example.com/goooo/gooo"},
			setup: func(t *testing.T) {
				os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq"), 0755)
			},
			expectErr: true,
		},
		{
			name:  "permission denied",
			input: []string{"rm", "motemen/ghqq"},
			setup: func(t *testing.T) {
				f := filepath.Join(tmpd, "github.com", "motemen", "ghqq")
				os.MkdirAll(f, 0000)
				t.Cleanup(func() {
					os.Chmod(f, 0755)
				})
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipOnWin && runtime.GOOS == "windows" {
				t.SkipNow()
			}

			if tc.setup != nil {
				tc.setup(t)
			}

			cmdutil.CommandRunner = commandRunner
			if tc.cmdRun != nil {
				cmdutil.CommandRunner = tc.cmdRun
			}

			var runErr error
			_, _, err := captureWithInput([]string{"y"}, func() {
				a := newApp()
				args := append([]string{"ghq"}, tc.input...)
				runErr = a.Run(context.Background(), args)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotErr := runErr != nil; gotErr != tc.expectErr {
				t.Fatalf("error = %v, expectErr = %v", runErr, tc.expectErr)
			}
		})
	}
}

func TestRmMultipleRepositories(t *testing.T) {
	root := newTempDir(t)
	setRepositoryPathTestRoot(t, root)
	names := []string{"motemen/ghq-first", "motemen/ghq-second"}
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(root, "github.com", filepath.FromSlash(name), ".git"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	var runErr error
	_, _, err := captureWithInput([]string{"y", "y"}, func() {
		runErr = newApp().Run(context.Background(), append([]string{"ghq", "rm"}, names...))
	})
	if err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatal(runErr)
	}
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, "github.com", filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Errorf("repository %s was not removed: %v", name, err)
		}
	}
}

func TestRmDryRunCommand(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	commandRunner := func(cmd *exec.Cmd) error {
		return nil
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpd := newTempDir(t)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpd)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	testCases := []struct {
		name      string
		input     []string
		setup     func(t *testing.T)
		expectErr bool
		cmdRun    func(cmd *exec.Cmd) error
		skipOnWin bool
	}{
		{
			name:  "simple",
			input: []string{"rm", "--dry-run", "motemen/ghqq"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: false,
		},
		{
			name:      "empty directory",
			input:     []string{"rm", "--dry-run", "motemen/ghqqq"},
			setup:     func(t *testing.T) {},
			expectErr: true,
		},
		{
			name:  "incorrect repository name",
			input: []string{"rm", "--dry-run", "example.com/goooo/gooo"},
			setup: func(t *testing.T) {
				os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq"), 0755)
			},
			expectErr: true,
		},
		{
			name:  "permission denied",
			input: []string{"rm", "--dry-run", "motemen/ghq-notpermitted"},
			setup: func(t *testing.T) {
				f := filepath.Join(tmpd, "github.com", "motemen", "ghq-notpermitted")
				os.MkdirAll(f, 0000)
				t.Cleanup(func() {
					os.Chmod(f, 0755)
				})
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipOnWin && runtime.GOOS == "windows" {
				t.SkipNow()
			}

			if tc.setup != nil {
				tc.setup(t)
			}

			cmdutil.CommandRunner = commandRunner
			if tc.cmdRun != nil {
				cmdutil.CommandRunner = tc.cmdRun
			}

			var runErr error
			_, _, err := capture(func() {
				a := newApp()
				args := append([]string{"ghq"}, tc.input...)
				runErr = a.Run(context.Background(), args)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotErr := runErr != nil; gotErr != tc.expectErr {
				t.Fatalf("error = %v, expectErr = %v", runErr, tc.expectErr)
			}
		})
	}
}

func TestRmWorktree(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error { return nil }

	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpd := newTempDir(t)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpd)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	t.Run("rm_linked_worktree", func(t *testing.T) {
		// Create main repo inside ghq root
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-rm", "main"),
			"https://github.com/wt-rm/main.git")

		// Create worktree registered under ghq root so ghq rm can resolve it
		wtDir := filepath.Join(tmpd, "github.com", "wt-rm", "wt-linked")
		addWorktree(t, mainDir, wtDir, "wt-rm-branch")

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-rm/wt-linked"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Worktree directory should be gone
		if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
			t.Error("worktree directory should be removed")
		}

		// Parent repo's .git/worktrees/<name> should be cleaned up
		wtEntry := filepath.Join(mainDir, ".git", "worktrees", "wt-linked")
		if _, err := os.Stat(wtEntry); !os.IsNotExist(err) {
			t.Error("parent repo's worktree entry should be cleaned up")
		}

		// Parent repo should still work
		c := exec.Command("git", "status")
		c.Dir = mainDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Errorf("git status in parent repo failed: %v\n%s", err, out)
		}
	})

	t.Run("rm_dryrun_worktree", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-dry", "main"),
			"https://github.com/wt-dry/main.git")
		wtDir := filepath.Join(tmpd, "github.com", "wt-dry", "wt-linked")
		addWorktree(t, mainDir, wtDir, "wt-dry-branch")

		out, _, err := capture(func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "--dry-run", "wt-dry/wt-linked"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Would remove worktree") {
			t.Errorf("expected 'Would remove worktree' in output, got: %s", out)
		}
		if _, err := os.Stat(wtDir); os.IsNotExist(err) {
			t.Error("worktree should still exist after dry-run")
		}
	})

	t.Run("rm_repo_with_linked_worktrees", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-parent", "repo"),
			"https://github.com/wt-parent/repo.git")

		// Create two external worktrees
		wt1 := filepath.Join(tmpd, "external-wt1")
		wt2 := filepath.Join(tmpd, "external-wt2")
		addWorktree(t, mainDir, wt1, "branch1")
		addWorktree(t, mainDir, wt2, "branch2")

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-parent/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Main repo should be gone
		if _, err := os.Stat(mainDir); !os.IsNotExist(err) {
			t.Error("main repo should be removed")
		}

		// Both worktree directories should be gone
		if _, err := os.Stat(wt1); !os.IsNotExist(err) {
			t.Error("worktree 1 should be removed")
		}
		if _, err := os.Stat(wt2); !os.IsNotExist(err) {
			t.Error("worktree 2 should be removed")
		}
	})

	t.Run("rm_dryrun_with_linked_worktrees", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-dry2", "repo"),
			"https://github.com/wt-dry2/repo.git")

		wt1 := filepath.Join(tmpd, "dry-wt1")
		addWorktree(t, mainDir, wt1, "dry-branch1")

		out, _, err := capture(func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "--dry-run", "wt-dry2/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "linked worktree") {
			t.Errorf("expected 'linked worktree' in output, got: %s", out)
		}
		if _, err := os.Stat(mainDir); os.IsNotExist(err) {
			t.Error("repo should still exist after dry-run")
		}
	})

	t.Run("rm_repo_with_already_deleted_worktree", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-gone", "repo"),
			"https://github.com/wt-gone/repo.git")

		wt := filepath.Join(tmpd, "gone-wt")
		addWorktree(t, mainDir, wt, "gone-branch")

		// Manually delete the worktree directory (simulating user deleting it)
		os.RemoveAll(wt)

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-gone/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Main repo should be gone regardless
		if _, err := os.Stat(mainDir); !os.IsNotExist(err) {
			t.Error("main repo should be removed even with pre-deleted worktree")
		}
	})
}

func TestRmBareWorktrees(t *testing.T) {
	for _, layout := range []string{"classic", "clean", "normal.git", "clean.git"} {
		t.Run(layout, func(t *testing.T) {
			tmp := newTempDir(t)
			root := filepath.Join(tmp, "root")
			setRepositoryPathTestRoot(t, root)
			t.Setenv("GIT_TEMPLATE_DIR", t.TempDir())
			seed := initGitRepo(t, filepath.Join(tmp, "seed"), "https://github.com/bare-rm/repo.git")
			repoDir := filepath.Join(root, "github.com", "bare-rm", "repo")
			args := []string{"bare-rm/repo"}
			if layout != "clean" {
				repoDir += ".git"
				args = append([]string{"--bare"}, args...)
			}
			cloneDir, cloneArgs := filepath.Join(repoDir, ".git"), []string{"clone", "--bare"}
			switch layout {
			case "classic":
				cloneDir = repoDir
			case "normal.git":
				cloneDir, cloneArgs = repoDir, []string{"clone"}
			}
			if out, err := exec.Command("git", append(cloneArgs, seed, cloneDir)...).CombinedOutput(); err != nil {
				t.Fatalf("git clone: %v\n%s", err, out)
			}
			worktrees := []string{filepath.Join(tmp, "external-one"), filepath.Join(tmp, "external two")}
			addWorktree(t, repoDir, worktrees[0], "first")
			addWorktree(t, repoDir, worktrees[1], "second")
			t.Setenv("GIT_CONFIG_COUNT", "1")
			t.Setenv("GIT_CONFIG_KEY_0", "safe.bareRepository")
			t.Setenv("GIT_CONFIG_VALUE_0", "explicit")
			out, _, err := runRm(t, "", append([]string{"--dry-run"}, args...)...)
			if err != nil || !strings.Contains(out, "2 linked worktree(s)") {
				t.Fatalf("dry-run: %v, %q", err, out)
			}
			for _, wt := range worktrees {
				if !strings.Contains(out, wt) {
					t.Errorf("dry-run omitted %s: %s", wt, out)
				}
				if _, err := os.Stat(wt); err != nil {
					t.Fatal(err)
				}
			}
			_, prompt, err := runRm(t, "n", args...)
			if err == nil || err.Error() != "aborted" || !strings.Contains(prompt, "2 linked worktree(s)") {
				t.Fatalf("cancellation: %v, %q", err, prompt)
			}
			for _, wt := range worktrees {
				if !strings.Contains(prompt, wt) {
					t.Errorf("confirmation omitted %s: %s", wt, prompt)
				}
				if out, err := exec.Command("git", "-C", wt, "status", "--porcelain").CombinedOutput(); err != nil {
					t.Fatalf("cancelled removal damaged %s: %v\n%s", wt, err, out)
				}
			}
			if _, _, err := runRm(t, "y", args...); err != nil {
				t.Fatal(err)
			}
			for _, path := range append(worktrees, repoDir) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("confirmed removal left %s: %v", path, err)
				}
			}
		})
	}
}

func TestRmSymlinkPreservesTargetWorktrees(t *testing.T) {
	testRmAliasPreservesTargetWorktrees(t, makeSymlink)
}

func testRmAliasPreservesTargetWorktrees(t *testing.T, makeAlias func(t *testing.T, target, alias string)) {
	t.Helper()
	for _, layout := range []string{"classic", "clean", "normal", "linked"} {
		t.Run(layout, func(t *testing.T) {
			tmp := newTempDir(t)
			primary, root := filepath.Join(tmp, "primary"), filepath.Join(tmp, "secondary")
			if err := os.Mkdir(primary, 0755); err != nil {
				t.Fatal(err)
			}
			setRepositoryPathTestRoot(t, primary+string(os.PathListSeparator)+root)
			t.Setenv("GIT_TEMPLATE_DIR", t.TempDir())
			seed := initGitRepo(t, filepath.Join(tmp, "seed"), "https://github.com/alias-rm/repo.git")
			repoDir, target := seed, seed
			alias := filepath.Join(root, "github.com", "alias-rm", "repo")
			args := []string{"alias-rm/repo"}
			if layout == "classic" || layout == "clean" {
				repoDir = filepath.Join(tmp, "target")
				gitDir := filepath.Join(repoDir, ".git")
				if layout == "classic" {
					repoDir += ".git"
					gitDir, alias = repoDir, alias+".git"
					args = append([]string{"--bare"}, args...)
				}
				target = repoDir
				if out, err := exec.Command("git", "clone", "--bare", seed, gitDir).CombinedOutput(); err != nil {
					t.Fatalf("git clone --bare: %v\n%s", err, out)
				}
			}
			wt := filepath.Join(tmp, "worktree with spaces")
			addWorktree(t, repoDir, wt, "first")
			if layout == "linked" {
				target = wt
			}
			if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
				t.Fatal(err)
			}
			makeAlias(t, target, alias)
			discovered, err := LocalRepositoryFromURL(mustParseURL("https://github.com/alias-rm/repo"), bareModeFromClassicBool(layout == "classic"))
			if err != nil || discovered.FullPath != alias {
				t.Fatalf("secondary-root alias not discovered: %v, %v", discovered, err)
			}
			marker := filepath.Join(wt, "keep")
			if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			registered, err := exec.Command("git", "-C", repoDir, "worktree", "list", "--porcelain").CombinedOutput()
			if err != nil {
				t.Fatalf("git worktree list: %v\n%s", err, registered)
			}
			checkRmAlias(t, alias, args, func() {
				if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
					t.Errorf("worktree contents changed: %v", err)
				}
				if out, err := exec.Command("git", "-C", wt, "status", "--porcelain").CombinedOutput(); err != nil {
					t.Errorf("worktree no longer usable: %v\n%s", err, out)
				}
				out, err := exec.Command("git", "-C", repoDir, "worktree", "list", "--porcelain").CombinedOutput()
				if err != nil || string(out) != string(registered) {
					t.Errorf("worktree registrations changed: %v\n%s", err, out)
				}
			})
		})
	}
}

func TestRmSymlinkWithoutReadableTarget(t *testing.T) {
	testRmAliasWithoutReadableTarget(t, makeSymlink)
}

func testRmAliasWithoutReadableTarget(t *testing.T, makeAlias func(t *testing.T, target, alias string)) {
	t.Helper()
	for _, state := range []string{"dangling", "empty", "unreadable"} {
		t.Run(state, func(t *testing.T) {
			if state == "unreadable" && runtime.GOOS == "windows" {
				t.Skip("chmod does not enforce directory read permissions on Windows")
			}
			tmp := newTempDir(t)
			root, target := filepath.Join(tmp, "root"), filepath.Join(tmp, "target")
			setRepositoryPathTestRoot(t, root)
			if err := os.Mkdir(target, 0755); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(root, "github.com", "alias-rm", "repo")
			if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
				t.Fatal(err)
			}
			makeAlias(t, target, alias)
			switch state {
			case "dangling":
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if err := os.Chmod(target, 0000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Chmod(target, 0755) })
				if _, err := os.ReadDir(target); !os.IsPermission(err) {
					t.Skipf("cannot enforce unreadable target: %v", err)
				}
			}
			before, err := os.Lstat(target)
			if state != "dangling" && err != nil {
				t.Fatal(err)
			}
			checkRmAlias(t, alias, []string{"alias-rm/repo"}, func() {
				after, err := os.Lstat(target)
				if state == "dangling" {
					if !os.IsNotExist(err) {
						t.Errorf("dangling target changed: %v", err)
					}
				} else if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
					t.Errorf("target changed: %v", err)
				}
			})
		})
	}
}

func TestRmSymlinkRootLookup(t *testing.T) {
	testRmAliasRootLookup(t, makeSymlink)
}

func TestRmDiscoveredRepository(t *testing.T) {
	tmp := newTempDir(t)
	primary, secondary := filepath.Join(tmp, "primary"), filepath.Join(tmp, "secondary")
	setRepositoryPathTestRoot(t, primary+string(os.PathListSeparator)+secondary)
	t.Setenv("GIT_TEMPLATE_DIR", t.TempDir())
	rel := filepath.Join("github.com", "alias-rm", "repo")
	repo := initGitRepo(t, filepath.Join(secondary, rel), "https://github.com/alias-rm/repo.git")
	alias := filepath.Join(primary, rel)
	if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, filepath.Join(tmp, "missing"), alias)
	out, _, err := runRm(t, "", "--dry-run", "alias-rm/repo")
	if err != nil || out != "Would remove "+repo+"\n" {
		t.Fatalf("discovered repository was replaced by an alias: %v, %q", err, out)
	}
}

func testRmAliasRootLookup(t *testing.T, makeAlias func(t *testing.T, target, alias string)) {
	t.Helper()
	for _, tc := range []struct {
		name, entry  string
		middle, bare bool
	}{
		{"empty-primary", "empty", false, false},
		{"nonempty-primary", "nonempty", false, false},
		{"file-primary", "file", false, false},
		{"file-middle-bare", "file", true, true},
		{"unreadable-primary", "unreadable", false, false},
		{"unreadable-middle", "unreadable", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.entry == "unreadable" && runtime.GOOS == "windows" {
				t.Skip("chmod does not enforce directory read permissions on Windows")
			}
			tmp := newTempDir(t)
			primary, secondary := filepath.Join(tmp, "primary"), filepath.Join(tmp, "secondary")
			roots, obstacle := []string{primary, secondary}, primary
			if tc.middle {
				obstacle = filepath.Join(tmp, "middle")
				roots = []string{primary, obstacle, secondary}
			}
			if tc.entry == "empty" || tc.entry == "nonempty" {
				obstacle = filepath.Join(primary, "github.com", "alias-rm", "repo")
			}
			if err := os.MkdirAll(obstacle, 0755); err != nil {
				t.Fatal(err)
			}
			marker := ""
			switch tc.entry {
			case "nonempty":
				marker = filepath.Join(obstacle, "keep")
			case "file":
				obstacle = filepath.Join(obstacle, "github.com")
				marker = obstacle
			case "unreadable":
				if err := os.Chmod(obstacle, 0000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Chmod(obstacle, 0755) })
				if _, err := os.Lstat(filepath.Join(obstacle, "github.com")); !os.IsPermission(err) {
					t.Skipf("cannot enforce inaccessible root: %v", err)
				}
			}
			if marker != "" {
				if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.Lstat(obstacle)
			if err != nil {
				t.Fatal(err)
			}
			setRepositoryPathTestRoot(t, strings.Join(roots, string(os.PathListSeparator)))
			alias := filepath.Join(secondary, "github.com", "alias-rm", "repo")
			args := []string{"alias-rm/repo"}
			if tc.bare {
				alias += ".git"
				args = append([]string{"--bare"}, args...)
			}
			if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(tmp, "target")
			if err := os.Mkdir(target, 0755); err != nil {
				t.Fatal(err)
			}
			makeAlias(t, target, alias)
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			selected, isAlias, err := resolveRemovalEntry(mustParseURL("https://github.com/alias-rm/repo"), bareModeFromClassicBool(tc.bare))
			if err != nil || !isAlias || selected != alias {
				t.Fatalf("alias lookup: path=%q, alias=%v, err=%v", selected, isAlias, err)
			}
			if _, _, err := runRm(t, "y", args...); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(alias); !os.IsNotExist(err) {
				t.Errorf("alias was not removed: %v", err)
			}
			if after, err := os.Lstat(obstacle); err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
				t.Errorf("unrelated root entry changed: %v", err)
			}
			if marker != "" {
				if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
					t.Errorf("unrelated contents changed: %v", err)
				}
			}
		})
	}
}

func makeSymlink(t *testing.T, target, alias string) {
	t.Helper()
	if err := os.Symlink(target, alias); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
}

func TestRmSymlinkReplacedDuringConfirmation(t *testing.T) {
	testRmAliasReplacedDuringConfirmation(t, makeSymlink)
}

func TestRmSymlinkURLRoot(t *testing.T) {
	testRmAliasURLRoot(t, makeSymlink)
}

func testRmAliasURLRoot(t *testing.T, makeAlias func(t *testing.T, target, alias string)) {
	t.Helper()
	tmp := newTempDir(t)
	primary, specific := filepath.Join(tmp, "primary"), filepath.Join(tmp, "specific")
	setRepositoryPathTestRoot(t, "")
	t.Cleanup(gitconfig.WithConfig(t, fmt.Sprintf("[ghq]\nroot = %q\n[ghq \"https://github.com/alias-rm/repo\"]\nroot = %q\n", primary, specific)))
	args := []string{"--bare", "https://github.com/alias-rm/repo"}
	rel := filepath.Join("github.com", "alias-rm", "repo.git")
	target := filepath.Join(tmp, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{primary, specific} {
		alias := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
			t.Fatal(err)
		}
		makeAlias(t, target, alias)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	out, _, err := runRm(t, "", append([]string{"--dry-run"}, args...)...)
	if err != nil || out != "Would remove "+filepath.Join(specific, rel)+"\n" {
		t.Fatalf("URL root was not preserved: %v, %q", err, out)
	}
	if _, _, err := runRm(t, "y", args...); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(primary, rel)); err != nil {
		t.Errorf("primary alias changed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(specific, rel)); !os.IsNotExist(err) {
		t.Errorf("URL-specific alias was not removed: %v", err)
	}
}

func testRmAliasReplacedDuringConfirmation(t *testing.T, makeAlias func(t *testing.T, target, alias string)) {
	t.Helper()
	tmp := newTempDir(t)
	root, target := filepath.Join(tmp, "root"), filepath.Join(tmp, "target")
	setRepositoryPathTestRoot(t, root)
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "github.com", "alias-rm", "repo")
	if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
		t.Fatal(err)
	}
	makeAlias(t, target, alias)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	input, answer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer answer.Close()
	promptReader, promptWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer promptReader.Close()
	defer promptWriter.Close()
	stdin, stderr := os.Stdin, os.Stderr
	os.Stdin, os.Stderr = input, promptWriter
	defer func() { os.Stdin, os.Stderr = stdin, stderr }()
	app := newApp()
	app.Writer = io.Discard
	done := make(chan error, 1)
	go func() {
		runErr := app.Run(context.Background(), []string{"ghq", "rm", "alias-rm/repo"})
		promptWriter.Close()
		done <- runErr
	}()
	expectedPrompt := "Remove " + alias + "? [y/N]: "
	prompt := make([]byte, len(expectedPrompt))
	_, promptErr := io.ReadFull(promptReader, prompt)
	marker := filepath.Join(alias, "keep")
	if promptErr == nil {
		if err := os.Remove(alias); err != nil {
			t.Error(err)
		} else if err := os.Mkdir(alias, 0755); err != nil {
			t.Error(err)
		} else if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
			t.Error(err)
		}
	}
	if _, err := fmt.Fprintln(answer, "y"); err != nil {
		t.Error(err)
	}
	answer.Close()
	runErr := <-done
	if promptErr != nil || string(prompt) != expectedPrompt {
		t.Errorf("confirmation: %v, got %q", promptErr, prompt)
	}
	if runErr == nil {
		t.Error("removal accepted a nonempty replacement directory")
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
		t.Errorf("replacement contents were deleted: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("dangling target changed: %v", err)
	}
}

func TestRmRegularFilePreserved(t *testing.T) {
	root := newTempDir(t)
	setRepositoryPathTestRoot(t, root)
	path := filepath.Join(root, "github.com", "file-rm", "repo")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	_, prompt, err := runRm(t, "y", "file-rm/repo")
	if err == nil || strings.Contains(prompt, "[y/N]") {
		t.Errorf("regular file not rejected before confirmation: %v, %q", err, prompt)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "keep" {
		t.Errorf("regular file changed: %v", err)
	}
}

func runRm(t *testing.T, answer string, args ...string) (string, string, error) {
	t.Helper()
	var runErr error
	out, prompt, err := captureWithInput([]string{answer}, func() {
		runErr = newApp().Run(context.Background(), append([]string{"ghq", "rm"}, args...))
	})
	if err != nil {
		t.Fatal(err)
	}
	return out, prompt, runErr
}

func checkRmAlias(t *testing.T, alias string, args []string, checkTarget func()) {
	t.Helper()
	out, _, err := runRm(t, "", append([]string{"--dry-run"}, args...)...)
	if err != nil || out != "Would remove "+alias+"\n" {
		t.Fatalf("dry-run: %v, %q", err, out)
	}
	checkTarget()
	_, prompt, err := runRm(t, "n", args...)
	if err == nil || err.Error() != "aborted" || prompt != "Remove "+alias+"? [y/N]: " {
		t.Errorf("cancellation: %v, %q", err, prompt)
	}
	if _, err := os.Readlink(alias); err != nil {
		t.Errorf("cancellation changed alias: %v", err)
	}
	checkTarget()
	out, prompt, err = runRm(t, "y", args...)
	if err != nil || out != "Removed "+alias+"\n" || prompt != "Remove "+alias+"? [y/N]: " {
		t.Errorf("removal: %v, %q", err, prompt)
	}
	if _, err := os.Lstat(alias); !os.IsNotExist(err) {
		t.Errorf("alias was not removed: %v", err)
	}
	checkTarget()
}

func TestRmRepositoryPathCollapse(t *testing.T) {
	type testCase struct {
		ref, path string
		bare      bool
	}
	cases := []testCase{
		{ref: "https://github.com", path: "github.com"},
		{ref: "https://example.com", path: "example.com"},
		{ref: "https://github.com/owner", path: "github.com/owner"},
		{ref: "https://github.com:443/owner", path: "github.com/owner"},
		{ref: "https://GITHUB.COM/owner", path: "GITHUB.COM/owner"},
		{ref: "https://github.com./owner", path: "github.com./owner"},
		{ref: "https://github.com/.git", path: "github.com"},
		{ref: "https://github.com/owner/.git", path: "github.com/owner.git", bare: true},
		{ref: "https://github.com/../../outside", path: "../outside"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases, testCase{ref: "https://github.com/%2e%2e%20/%2e%2e%20/outside", path: "../outside"})
	}
	for _, tc := range cases {
		t.Run(tc.ref, func(t *testing.T) {
			root := filepath.Join(newTempDir(t), "root")
			setRepositoryPathTestRoot(t, root)
			marker := filepath.Join(root, filepath.FromSlash(tc.path), "repo", "keep")
			if err := os.MkdirAll(filepath.Dir(marker), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			args := []string{"ghq", "rm"}
			if tc.bare {
				args = append(args, "--bare")
			}
			var runErr error
			_, stderr, err := captureWithInput([]string{"y"}, func() {
				runErr = newApp().Run(context.Background(), append(args, tc.ref))
			})
			if err != nil {
				t.Fatal(err)
			}
			if runErr == nil || !strings.Contains(runErr.Error(), "repository") {
				t.Errorf("expected repository rejection, got %v", runErr)
			}
			if strings.Contains(stderr, "[y/N]") {
				t.Error("requested confirmation for a collapsed path")
			}
			if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
				t.Errorf("existing repository was changed: %q, %v", data, err)
			}
		})
	}
}
