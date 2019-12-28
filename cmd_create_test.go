package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/Songmu/gitconfig"
	"github.com/motemen/ghq/cmdutil"
)

func TestDoCreate(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	var lastCmd *exec.Cmd
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		lastCmd = cmd
		return nil
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	defer gitconfig.WithConfig(t, "")()
	tmpd := newTempDir(t)
	defer os.RemoveAll(tmpd)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	defer tmpEnv(envGhqRoot, tmpd)()
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	out, _, _ := capture(func() {
		newApp().Run([]string{"", "create", "motemen/ghqq"})
	})
	out = strings.TrimSpace(out)
	wantDir := filepath.Join(tmpd, "github.com/motemen/ghqq")

	wantArgs := []string{"git", "init"}
	if !reflect.DeepEqual(lastCmd.Args, wantArgs) {
		t.Errorf("cmd.Args = %v, want: %v", lastCmd.Args, wantArgs)
	}

	if lastCmd.Dir != wantDir {
		t.Errorf("cmd.Dir = %q, want: %q", lastCmd.Dir, wantDir)
	}

	if out != wantDir {
		t.Errorf("cmd.Dir = %q, want: %q", out, wantDir)
	}
}
