package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Songmu/gitconfig"
	"github.com/motemen/ghq/cmdutil"
)

func TestDoExpand(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	defer func(orig string) { _home = orig }(_home)
	defer gitconfig.WithConfig(t, "")()
	tmpd := newTempDir(t)
	defer os.RemoveAll(tmpd)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	_localRepositoryRoots = []string{tmpd}

	out, _, _ := capture(func() {
		newApp().Run([]string{"", "expand", "motemen/ghqq"})
	})
	out = strings.TrimSpace(out)
	wantDir := filepath.Join(tmpd, "github.com/motemen/ghqq")

	if out != wantDir {
		t.Errorf("cmd.Dir = %q, want: %q", out, wantDir)
	}
}
