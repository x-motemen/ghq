package main

import (
"os"
"os/exec"
"path/filepath"
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
