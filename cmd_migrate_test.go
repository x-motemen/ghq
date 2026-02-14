package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

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
		e := a.Run([]string{"ghq", "migrate", "-y", srcdir})
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
		e := a.Run([]string{"ghq", "migrate", "-y", "/does/not/exist"})
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
		e := a.Run([]string{"ghq", "migrate", "--dry-run", srcdir})
		if e != nil {
			t.Fatal(e)
		}

		if _, err := os.Stat(srcdir); os.IsNotExist(err) {
			t.Error("source should still exist")
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
		e := a.Run([]string{"ghq", "migrate", "-y", srcdir})
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
		e := a.Run([]string{"ghq", "migrate", "-y", srcdir})
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
		e := a.Run([]string{"ghq", "migrate", "-y", srcdir})
		if e == nil {
			t.Error("should fail when dest exists")
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
		e := a.Run([]string{"ghq", "migrate", "-y", srcdir})
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
