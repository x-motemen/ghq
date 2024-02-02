package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

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
				os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq"), 0755)
			},
			expectErr: false,
		},
		{
			name:      "empty directory",
			input:     []string{"rm", "motemen/ghqqq"},
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
		})
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
				os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq"), 0755)
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
			input: []string{"rm", "--dry-run", "motemen/ghqq"},
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
		})
	}
}
