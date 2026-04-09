package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/x-motemen/ghq/cmdutil"
)

func TestDoRm(t *testing.T) {
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

	t.Run("rm_regular_repo", func(t *testing.T) {
		repoDir := filepath.Join(tmpd, "github.com", "motemen", "ghqq")
		os.MkdirAll(repoDir, 0755)
		os.WriteFile(filepath.Join(repoDir, ".git"), []byte(""), 0644)

		out, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "motemen/ghqq"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Removed") {
			t.Errorf("expected 'Removed' in output, got: %s", out)
		}
		if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
			t.Error("repo directory should be removed")
		}
	})

	t.Run("rm_nonexistent", func(t *testing.T) {
		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "rm", "motemen/doesnotexist"})
		if e == nil {
			t.Error("expected error for nonexistent repo")
		}
	})

	t.Run("rm_dryrun", func(t *testing.T) {
		repoDir := filepath.Join(tmpd, "github.com", "motemen", "dryrm")
		os.MkdirAll(repoDir, 0755)
		os.WriteFile(filepath.Join(repoDir, ".git"), []byte(""), 0644)

		out, _, err := capture(func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "--dry-run", "motemen/dryrm"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Would remove") {
			t.Errorf("expected 'Would remove' in output, got: %s", out)
		}
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			t.Error("repo should still exist after dry-run")
		}
	})

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
