package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
)

// initGitRepo creates a git repo at dir with the given remote URL and an
// initial empty commit. It returns dir for convenience.
func initGitRepo(t *testing.T, dir, remoteURL string) string {
	t.Helper()
	os.MkdirAll(dir, 0755)

	for _, args := range [][]string{
		{"init"},
		{"remote", "add", "origin", remoteURL},
		{"-c", "user.name=test", "-c", "user.email=test@test.com",
			"-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", args[0], err, out)
		}
	}
	return dir
}

// addWorktree creates a git worktree at wtDir branching from the repo at repoDir.
func addWorktree(t *testing.T, repoDir, wtDir, branch string) {
	t.Helper()
	c := exec.Command("git", "worktree", "add", "-b", branch, wtDir)
	c.Dir = repoDir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}
}

// Test for the migrate command
func TestDoMigrate(t *testing.T) {
	defer func(x string) { _home = x }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpdir := newTempDir(t)
	defer func(y []string) { _localRepositoryRoots = y }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpdir)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	// Test case: successful migration
	t.Run("migrate_success", func(t *testing.T) {
		srcdir := filepath.Join(tmpdir, "sources", "proj")
		os.MkdirAll(srcdir, 0755)

		c1 := exec.Command("git", "init")
		c1.Dir = srcdir
		c1.Run()

		c2 := exec.Command("git", "remote", "add", "origin", "https://github.com/alice/proj.git")
		c2.Dir = srcdir
		c2.Run()

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e != nil {
			t.Fatal(e)
		}

		dest := filepath.Join(tmpdir, "github.com", "alice", "proj")
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Error("dest not found")
		}
	})

	// Test case: nonexistent directory
	t.Run("migrate_nonexist", func(t *testing.T) {
		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", "/does/not/exist"})
		if e == nil {
			t.Error("expected error")
		}
	})

	// Test case: dry run
	t.Run("migrate_dryrun", func(t *testing.T) {
		srcdir := filepath.Join(tmpdir, "sources2", "proj2")
		os.MkdirAll(srcdir, 0755)

		c1 := exec.Command("git", "init")
		c1.Dir = srcdir
		c1.Run()

		c2 := exec.Command("git", "remote", "add", "origin", "https://github.com/bob/proj2.git")
		c2.Dir = srcdir
		c2.Run()

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "--dry-run", srcdir})
		if e != nil {
			t.Fatal(e)
		}

		if _, err := os.Stat(srcdir); os.IsNotExist(err) {
			t.Error("source should still exist")
		}
	})

	// Test case: migrate repo with linked worktrees repairs forward references
	t.Run("migrate_with_linked_worktrees", func(t *testing.T) {
		srcdir := initGitRepo(t, filepath.Join(tmpdir, "sources_wt", "main"),
			"https://github.com/wt-user/main.git")
		wtDir := filepath.Join(tmpdir, "sources_wt", "wt")
		addWorktree(t, srcdir, wtDir, "wt-branch")

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e != nil {
			t.Fatal(e)
		}

		dest := filepath.Join(tmpdir, "github.com", "wt-user", "main")
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Error("dest not found")
		}

		// Verify worktree's .git file has exact gitdir: reference to new location
		content, err := os.ReadFile(filepath.Join(wtDir, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		wantGitdir := "gitdir: " + filepath.ToSlash(filepath.Join(dest, ".git", "worktrees", "wt"))
		if got := strings.TrimSpace(string(content)); got != wantGitdir {
			t.Errorf("worktree .git:\n  got:  %s\n  want: %s", got, wantGitdir)
		}

		// Verify git status works in the worktree after migration
		c := exec.Command("git", "status")
		c.Dir = wtDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Errorf("git status in worktree failed after migration: %v\n%s", err, out)
		}
	})

	// Test case: dry run with linked worktrees mentions repair
	t.Run("migrate_dryrun_with_worktrees", func(t *testing.T) {
		srcdir := initGitRepo(t, filepath.Join(tmpdir, "sources_wt_dry", "main"),
			"https://github.com/wt-dry/proj.git")
		wtDir := filepath.Join(tmpdir, "sources_wt_dry", "wt")
		addWorktree(t, srcdir, wtDir, "wt-dry-branch")

		out, _, err := capture(func() {
			a := newApp()
			a.Run(context.Background(), []string{"ghq", "migrate", "--dry-run", srcdir})
		})
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(out, "Would migrate") {
			t.Errorf("expected dry-run migration message, got: %s", out)
		}
		if !strings.Contains(out, "worktree repair") {
			t.Errorf("expected worktree repair mention in dry-run, got: %s", out)
		}
		if _, err := os.Stat(srcdir); os.IsNotExist(err) {
			t.Error("source should still exist in dry-run mode")
		}
	})

	// Test case: worktree inside the repo directory moves along with it
	t.Run("migrate_with_internal_worktree", func(t *testing.T) {
		srcdir := initGitRepo(t, filepath.Join(tmpdir, "sources_wt_int", "main"),
			"https://github.com/wt-int/proj.git")
		// Create worktree INSIDE the repo directory
		wtDir := filepath.Join(srcdir, ".worktrees", "feat")
		addWorktree(t, srcdir, wtDir, "wt-int-branch")

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e != nil {
			t.Fatal(e)
		}

		dest := filepath.Join(tmpdir, "github.com", "wt-int", "proj")
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Error("dest not found")
		}

		// The worktree moved with the repo — verify its .git file
		// has exact gitdir: reference to the new main repo location
		newWtDir := filepath.Join(dest, ".worktrees", "feat")
		content, err := os.ReadFile(filepath.Join(newWtDir, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		wantGitdir := "gitdir: " + filepath.ToSlash(filepath.Join(dest, ".git", "worktrees", "feat"))
		if got := strings.TrimSpace(string(content)); got != wantGitdir {
			t.Errorf("internal worktree .git:\n  got:  %s\n  want: %s", got, wantGitdir)
		}

		// Verify git status works in the internal worktree after migration
		c := exec.Command("git", "status")
		c.Dir = newWtDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Errorf("git status in internal worktree failed after migration: %v\n%s", err, out)
		}
	})
}

func TestMigrateEdgeCases(t *testing.T) {
	defer func(x string) { _home = x }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpdir := newTempDir(t)
	defer func(y []string) { _localRepositoryRoots = y }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpdir)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	t.Run("no_vcs_backend", func(t *testing.T) {
		srcdir := filepath.Join(tmpdir, "src3", "not-repo")
		os.MkdirAll(srcdir, 0755)

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e == nil {
			t.Error("should fail when no VCS found")
		}
	})

	t.Run("no_remote_url", func(t *testing.T) {
		srcdir := filepath.Join(tmpdir, "src4", "no-rem")
		os.MkdirAll(srcdir, 0755)

		c := exec.Command("git", "init")
		c.Dir = srcdir
		c.Run()

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e == nil {
			t.Error("should fail when no remote")
		}
	})

	t.Run("dest_already_exists", func(t *testing.T) {
		srcdir := filepath.Join(tmpdir, "src5", "exist")
		os.MkdirAll(srcdir, 0755)

		c1 := exec.Command("git", "init")
		c1.Dir = srcdir
		c1.Run()

		c2 := exec.Command("git", "remote", "add", "origin", "https://github.com/user3/exist.git")
		c2.Dir = srcdir
		c2.Run()

		dest := filepath.Join(tmpdir, "github.com", "user3", "exist")
		os.MkdirAll(dest, 0755)

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e == nil {
			t.Error("should fail when dest exists")
		}
	})

	t.Run("migrate_worktree_refused", func(t *testing.T) {
		srcdir := initGitRepo(t, filepath.Join(tmpdir, "src_wt_ref", "main"),
			"https://github.com/wt-ref/proj.git")
		wtDir := filepath.Join(tmpdir, "src_wt_ref", "wt")
		addWorktree(t, srcdir, wtDir, "wt-ref-branch")

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", wtDir})
		if e == nil {
			t.Fatal("expected error migrating a worktree")
		}
		if !strings.Contains(e.Error(), "worktree or submodule") {
			t.Errorf("error should mention worktree or submodule, got: %v", e)
		}
		if !strings.Contains(e.Error(), ".git") {
			t.Errorf("error should mention .git link target, got: %v", e)
		}
	})

	t.Run("unsupported_vcs", func(t *testing.T) {
		// Create a CVS repository structure to test unsupported VCS
		srcdir := filepath.Join(tmpdir, "src6", "cvs-repo")
		cvsDir := filepath.Join(srcdir, "CVS")
		os.MkdirAll(cvsDir, 0755)

		// Create a minimal CVS/Repository file
		repoFile := filepath.Join(cvsDir, "Repository")
		os.WriteFile(repoFile, []byte("test-repo\n"), 0644)

		a := newApp()
		e := a.Run(context.Background(), []string{"ghq", "migrate", "-y", srcdir})
		if e == nil {
			t.Error("should fail for unsupported VCS (CVS)")
		}
		// Check that the error message mentions unsupported VCS
		if e != nil && !strings.Contains(e.Error(), "not supported") {
			t.Errorf("expected 'not supported' error, got: %v", e)
		}
	})
}

func TestMoveDir(t *testing.T) {
	tmpdir := newTempDir(t)

	t.Run("move_same_device", func(t *testing.T) {
		srcDir := filepath.Join(tmpdir, "move_src")
		dstDir := filepath.Join(tmpdir, "move_dst")

		os.MkdirAll(srcDir, 0755)
		os.WriteFile(filepath.Join(srcDir, "testfile.txt"), []byte("test"), 0644)

		if err := moveDir(srcDir, dstDir); err != nil {
			t.Fatal(err)
		}

		// Verify destination exists
		if _, err := os.Stat(dstDir); os.IsNotExist(err) {
			t.Error("destination directory does not exist")
		}

		// Verify source is gone
		if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
			t.Error("source directory still exists")
		}

		// Verify content
		content, err := os.ReadFile(filepath.Join(dstDir, "testfile.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "test" {
			t.Errorf("content mismatch: got %q, want %q", content, "test")
		}
	})

	t.Run("move_with_subdirectories", func(t *testing.T) {
		srcDir := filepath.Join(tmpdir, "move_src2")
		dstDir := filepath.Join(tmpdir, "move_dst2")

		os.MkdirAll(filepath.Join(srcDir, "sub1", "sub2"), 0755)
		os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root"), 0644)
		os.WriteFile(filepath.Join(srcDir, "sub1", "file1.txt"), []byte("file1"), 0644)
		os.WriteFile(filepath.Join(srcDir, "sub1", "sub2", "file2.txt"), []byte("file2"), 0644)

		if err := moveDir(srcDir, dstDir); err != nil {
			t.Fatal(err)
		}

		// Verify all files exist
		files := []string{
			"root.txt",
			"sub1/file1.txt",
			"sub1/sub2/file2.txt",
		}
		for _, f := range files {
			path := filepath.Join(dstDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("file %s does not exist", f)
			}
		}

		// Verify source is gone
		if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
			t.Error("source directory still exists")
		}
	})

	// Note: moveDir also has a cross-device (EXDEV) fallback path which is
	// difficult to exercise reliably in unit tests because it depends on
	// running across different filesystems. That behavior is validated in
	// higher-level integration tests / environments that provide multiple
	// mounts, rather than in this unit test.
}

func TestIsLinkedGitDir(t *testing.T) {
	tmpdir := newTempDir(t)

	t.Run("regular_repo", func(t *testing.T) {
		dir := filepath.Join(tmpdir, "regular")
		os.MkdirAll(dir, 0755)

		c := exec.Command("git", "init")
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}

		linked, _, err := isLinkedGitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if linked {
			t.Error("regular repo should not be detected as linked")
		}
	})

	t.Run("no_git", func(t *testing.T) {
		dir := filepath.Join(tmpdir, "nogit")
		os.MkdirAll(dir, 0755)

		linked, _, err := isLinkedGitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if linked {
			t.Error("directory without .git should not be detected as linked")
		}
	})

	t.Run("submodule_gitfile", func(t *testing.T) {
		dir := filepath.Join(tmpdir, "submod")
		os.MkdirAll(dir, 0755)

		// Simulate a submodule's .git file pointing to .git/modules/
		os.WriteFile(filepath.Join(dir, ".git"),
			[]byte("gitdir: ../.git/modules/submod\n"), 0644)

		linked, target, err := isLinkedGitDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !linked {
			t.Error("submodule should be detected as linked")
		}
		if !strings.Contains(target, "modules") {
			t.Errorf("target should reference modules dir, got: %s", target)
		}
	})

	t.Run("actual_worktree", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpdir, "wt_main"),
			"https://github.com/dummy/wt-main.git")
		wtDir := filepath.Join(tmpdir, "wt_linked")
		addWorktree(t, mainDir, wtDir, "wt-test")

		linked, target, err := isLinkedGitDir(wtDir)
		if err != nil {
			t.Fatal(err)
		}
		if !linked {
			t.Error("worktree should be detected as linked")
		}
		if !strings.Contains(target, "worktrees") {
			t.Errorf("target should reference worktrees dir, got: %s", target)
		}
	})
}

func TestDoMigrate_p4(t *testing.T) {
	tmpdir := newTempDir(t)
	setRepositoryPathTestRoot(t, tmpdir)
	srv := newP4FakeServer()
	calls := fakeP4(t, srv.respond)
	newSrc := func(name string) string {
		src := filepath.Join(tmpdir, "sources", name)
		os.MkdirAll(src, 0755)
		if err := writeP4Config(src, &p4Config{port: "p4.example.com:1666", client: "ws"}); err != nil {
			t.Fatal(err)
		}
		srv.clients["ws"] = p4TestSpec("ws", src, true, "//depot/proj/main/... //ws/...")
		return src
	}
	dest := filepath.Join(tmpdir, "p4.example.com", "depot", "proj", "main")

	t.Run("dry-run", func(t *testing.T) {
		src := newSrc("dry")
		out, _, _ := capture(func() {
			if err := newApp().Run(context.Background(), []string{"ghq", "migrate", "--dry-run", src}); err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(out, dest) || !strings.Contains(out, "p4 client") {
			t.Errorf("unexpected output: %s", out)
		}
		for _, c := range *calls {
			if slices.Contains(c.args, "-i") {
				t.Errorf("dry-run should not modify the client: %v", c.args)
			}
		}
	})

	t.Run("migrate", func(t *testing.T) {
		*calls = nil
		src := newSrc("proj")
		if err := newApp().Run(context.Background(), []string{"ghq", "migrate", "-y", src}); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dest, p4ConfigName)); err != nil {
			t.Fatal(err)
		}
		last := (*calls)[len(*calls)-1]
		if !reflect.DeepEqual(last.args, []string{"p4", "-p", "p4.example.com:1666", "-c", "ws", "client", "-i"}) ||
			!strings.Contains(last.stdin, "\nRoot:\t"+dest+"\n") {
			t.Errorf("client root should be updated, got %v:\n%s", last.args, last.stdin)
		}
	})
}

func TestDoMigrate_p4Review1(t *testing.T) {
	tmpdir := newTempDir(t)
	setRepositoryPathTestRoot(t, tmpdir)
	dest := filepath.Join(tmpdir, "p4.example.com", "depot", "proj", "main")
	newSrc := func(t *testing.T, name string) string {
		src := filepath.Join(tmpdir, "sources", name)
		os.MkdirAll(src, 0755)
		if err := writeP4Config(src, &p4Config{port: "p4.example.com:1666", client: "ws"}); err != nil {
			t.Fatal(err)
		}
		return src
	}
	saved := func(calls []p4Call) bool {
		for _, c := range calls {
			if slices.Contains(c.args, "-i") {
				return true
			}
		}
		return false
	}

	t.Run("copy of a workspace rooted elsewhere", func(t *testing.T) {
		src := newSrc(t, "copy")
		srv := newP4FakeServer()
		srv.clients["ws"] = p4TestSpec("ws", "/active/workspace", true, "//depot/proj/main/... //ws/...")
		calls := fakeP4(t, srv.respond)
		if err := newApp().Run(context.Background(), []string{"ghq", "migrate", "-y", src}); err == nil {
			t.Error("error should be occurred, but nil")
		}
		if _, err := os.Stat(filepath.Join(src, p4ConfigName)); err != nil {
			t.Errorf("source should stay: %v", err)
		}
		if saved(*calls) {
			t.Errorf("the other workspace's client must not be modified: %+v", *calls)
		}
	})

	t.Run("client root update fails", func(t *testing.T) {
		src := newSrc(t, "fail")
		srv := newP4FakeServer()
		srv.clients["ws"] = p4TestSpec("ws", src, true, "//depot/proj/main/... //ws/...")
		srv.fail["client -i"] = "Perforce password (P4PASSWD) invalid or unset."
		fakeP4(t, srv.respond)
		if err := newApp().Run(context.Background(), []string{"ghq", "migrate", "-y", src}); err == nil {
			t.Error("error should be occurred, but nil")
		}
		if _, err := os.Stat(filepath.Join(src, p4ConfigName)); err != nil {
			t.Errorf("workspace should be moved back: %v", err)
		}
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Errorf("destination should not be left behind: %v", err)
		}
	})
}
