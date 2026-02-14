package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// Test helper functions for cross-device migration
func TestCopyFile(t *testing.T) {
tmpdir := newTempDir(t)

t.Run("copy_regular_file", func(t *testing.T) {
srcFile := filepath.Join(tmpdir, "source.txt")
dstFile := filepath.Join(tmpdir, "dest.txt")
content := []byte("test content\n")

if err := os.WriteFile(srcFile, content, 0644); err != nil {
t.Fatal(err)
}

if err := copyFile(srcFile, dstFile, 0644); err != nil {
t.Fatal(err)
}

// Check destination exists and has correct content
dstContent, err := os.ReadFile(dstFile)
if err != nil {
t.Fatal(err)
}
if string(dstContent) != string(content) {
t.Errorf("content mismatch: got %q, want %q", dstContent, content)
}

// Check permissions (skip on Windows as it doesn't support Unix-style permissions)
if runtime.GOOS != "windows" {
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if dstInfo.Mode().Perm() != 0644 {
		t.Errorf("permission mismatch: got %o, want %o", dstInfo.Mode().Perm(), 0644)
	}
}
})

t.Run("copy_preserves_permissions", func(t *testing.T) {
srcFile := filepath.Join(tmpdir, "exec.sh")
dstFile := filepath.Join(tmpdir, "exec_copy.sh")

if err := os.WriteFile(srcFile, []byte("#!/bin/sh\n"), 0755); err != nil {
t.Fatal(err)
}

if err := copyFile(srcFile, dstFile, 0755); err != nil {
t.Fatal(err)
}

// Check permissions (skip on Windows as it doesn't support Unix-style permissions)
if runtime.GOOS != "windows" {
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if dstInfo.Mode().Perm() != 0755 {
		t.Errorf("permission mismatch: got %o, want %o", dstInfo.Mode().Perm(), 0755)
	}
}
})
}

func TestCopyDir(t *testing.T) {
tmpdir := newTempDir(t)

t.Run("copy_directory_tree", func(t *testing.T) {
srcDir := filepath.Join(tmpdir, "srcrepo")
dstDir := filepath.Join(tmpdir, "dstrepo")

// Create source directory structure
os.MkdirAll(filepath.Join(srcDir, "subdir1"), 0755)
os.MkdirAll(filepath.Join(srcDir, "subdir2", "nested"), 0755)
os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
os.WriteFile(filepath.Join(srcDir, "subdir1", "file2.txt"), []byte("content2"), 0644)
os.WriteFile(filepath.Join(srcDir, "subdir2", "nested", "file3.txt"), []byte("content3"), 0644)

if err := copyDir(srcDir, dstDir); err != nil {
t.Fatal(err)
}

// Verify all files and directories were copied
testCases := []struct {
path    string
content string
}{
{"file1.txt", "content1"},
{"subdir1/file2.txt", "content2"},
{"subdir2/nested/file3.txt", "content3"},
}

for _, tc := range testCases {
dstPath := filepath.Join(dstDir, tc.path)
content, err := os.ReadFile(dstPath)
if err != nil {
t.Errorf("failed to read %s: %v", tc.path, err)
continue
}
if string(content) != tc.content {
t.Errorf("%s: got %q, want %q", tc.path, content, tc.content)
}
}
})

t.Run("copy_with_symlinks", func(t *testing.T) {
srcDir := filepath.Join(tmpdir, "src_with_links")
dstDir := filepath.Join(tmpdir, "dst_with_links")

os.MkdirAll(srcDir, 0755)
targetFile := filepath.Join(srcDir, "target.txt")
linkFile := filepath.Join(srcDir, "link.txt")

os.WriteFile(targetFile, []byte("target content"), 0644)
if err := os.Symlink("target.txt", linkFile); err != nil {
t.Skip("symlink not supported on this platform")
}

if err := copyDir(srcDir, dstDir); err != nil {
t.Fatal(err)
}

// Verify symlink was copied as symlink
dstLink := filepath.Join(dstDir, "link.txt")
linkInfo, err := os.Lstat(dstLink)
if err != nil {
t.Fatal(err)
}
if linkInfo.Mode()&os.ModeSymlink == 0 {
t.Error("expected symlink, got regular file")
}

// Verify link target
linkTarget, err := os.Readlink(dstLink)
if err != nil {
t.Fatal(err)
}
if linkTarget != "target.txt" {
t.Errorf("link target: got %q, want %q", linkTarget, "target.txt")
}
})

t.Run("copy_preserves_directory_permissions", func(t *testing.T) {
srcDir := filepath.Join(tmpdir, "src_perms")
dstDir := filepath.Join(tmpdir, "dst_perms")

os.MkdirAll(srcDir, 0755)
subdir := filepath.Join(srcDir, "restricted")
os.MkdirAll(subdir, 0700)

if err := copyDir(srcDir, dstDir); err != nil {
t.Fatal(err)
}

// Check permissions (skip on Windows as it doesn't support Unix-style permissions)
if runtime.GOOS != "windows" {
	dstSubdir := filepath.Join(dstDir, "restricted")
	info, err := os.Stat(dstSubdir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("directory permission: got %o, want %o", info.Mode().Perm(), 0700)
	}
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
}
