package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/x-motemen/ghq/cmdutil"
)

func TestCommandCheck(t *testing.T) {
	tmpdir := newTempDir(t)
	t.Setenv("GHQ_ROOT", tmpdir)

	// Create a clean repository
	cleanRepoPath := filepath.Join(tmpdir, "github.com", "user", "clean")
	os.MkdirAll(cleanRepoPath, 0755)
	cmdutil.Run("git", "-C", cleanRepoPath, "init")
	cmdutil.Run("git", "-C", cleanRepoPath, "commit", "--allow-empty", "-m", "initial commit")

	// Create a repository with uncommitted changes
	dirtyRepoPath := filepath.Join(tmpdir, "github.com", "user", "dirty")
	os.MkdirAll(dirtyRepoPath, 0755)
	cmdutil.Run("git", "-C", dirtyRepoPath, "init")
	os.WriteFile(filepath.Join(dirtyRepoPath, "file.txt"), []byte("hello"), 0644)

	// Create a repository with a stash
	stashedRepoPath := filepath.Join(tmpdir, "github.com", "user", "stashed")
	os.MkdirAll(stashedRepoPath, 0755)
	cmdutil.Run("git", "-C", stashedRepoPath, "init")
	os.WriteFile(filepath.Join(stashedRepoPath, "file.txt"), []byte("hello"), 0644)
	cmdutil.Run("git", "-C", stashedRepoPath, "add", "file.txt")
	cmdutil.Run("git", "-C", stashedRepoPath, "stash")

	out, _, err := capture(func() {
		newApp().Run([]string{"ghq", "check"})
	})

	if err != nil {
		t.Errorf("error should be nil, but: %v", err)
	}

	if strings.Contains(out, "clean") {
		t.Errorf("clean repository should not be listed")
	}

	if !strings.Contains(out, "dirty") {
		t.Errorf("dirty repository should be listed")
	}

	if !strings.Contains(out, "stashed") {
		t.Errorf("stashed repository should be listed")
	}
}
