//go:build windows

package main

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestRmJunctionPreservesTargetWorktrees(t *testing.T) {
	testRmAliasPreservesTargetWorktrees(t, makeJunction)
}

func TestRmJunctionRootLookup(t *testing.T) {
	testRmAliasRootLookup(t, makeJunction)
}

func TestRmJunctionWithoutReadableTarget(t *testing.T) {
	testRmAliasWithoutReadableTarget(t, makeJunction)
}

func TestRmJunctionReplacedDuringConfirmation(t *testing.T) {
	testRmAliasReplacedDuringConfirmation(t, makeJunction)
}

func TestRmJunctionURLRoot(t *testing.T) {
	testRmAliasURLRoot(t, makeJunction)
}

func TestRmIrregularFilePreserved(t *testing.T) {
	const (
		fsctlSetReparsePoint    = 0x900A4
		fsctlDeleteReparsePoint = 0x900AC
	)
	root := newTempDir(t)
	setRepositoryPathTestRoot(t, root)
	path := filepath.Join(root, "github.com", "file-rm", "repo.git")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	contents := "keep"
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	// A private non-name-surrogate tag with an owner GUID models a non-link reparse file.
	reparse := make([]byte, 24)
	binary.LittleEndian.PutUint32(reparse, 0x42)
	copy(reparse[8:], "ghq-reparse-test")
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	var returned uint32
	err = syscall.DeviceIoControl(syscall.Handle(f.Fd()), fsctlSetReparsePoint, &reparse[0], uint32(len(reparse)), nil, 0, &returned, nil)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("set non-link reparse point: %v, close: %v", err, closeErr)
	}
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Type() != os.ModeIrregular {
		t.Fatalf("reparse file mode = %v, want a non-directory irregular file", fi.Mode())
	}
	if _, err := os.Readlink(path); err == nil {
		t.Fatal("reparse file was unexpectedly recognized as a link")
	}
	_, prompt, err := runRm(t, "y", "--bare", "file-rm/repo")
	if err == nil || strings.Contains(prompt, "[y/N]") {
		t.Errorf("non-link file was not rejected before confirmation: %v, %q", err, prompt)
	}
	if after, err := os.Lstat(path); err != nil || after.Mode().Type() != os.ModeIrregular {
		t.Fatalf("non-link reparse entry was changed: %v", err)
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = syscall.DeviceIoControl(handle, fsctlDeleteReparsePoint, &reparse[0], uint32(len(reparse)), nil, 0, &returned, nil)
	closeErr = syscall.CloseHandle(handle)
	if err != nil || closeErr != nil {
		t.Fatalf("clear reparse attribute: %v, close: %v", err, closeErr)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != contents {
		t.Errorf("non-link file contents were changed: %v", err)
	}
}

func makeJunction(t *testing.T, target, alias string) {
	t.Helper()
	out, err := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput()
	if err != nil {
		t.Fatalf("create junction: %v\n%s", err, out)
	}
}
