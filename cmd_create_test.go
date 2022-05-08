package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/x-motemen/ghq/cmdutil"
)

func TestDoCreate(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	var lastCmd *exec.Cmd
	commandRunner := func(cmd *exec.Cmd) error {
		lastCmd = cmd
		return nil
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpd := newTempDir(t)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	defer tmpEnv(envGhqRoot, tmpd)()
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	testCases := []struct {
		name      string
		input     []string
		want      []string
		wantDir   string
		errStr    string
		setup     func() func()
		cmdRun    func(cmd *exec.Cmd) error
		skipOnWin bool
	}{{
		name:    "simple",
		input:   []string{"create", "motemen/ghqq"},
		want:    []string{"git", "init"},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghqq"),
	}, {
		name:  "empty directory exists",
		input: []string{"create", "motemen/ghqqq"},
		want:  []string{"git", "init"},
		setup: func() func() {
			os.MkdirAll(filepath.Join(tmpd, "github.com/motemen/ghqqq"), 0755)
			return func() {}
		},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghqqq"),
	}, {
		name:  "invalid VCS",
		input: []string{"create", "example.com/goooo/gooo"},
		cmdRun: func(cmd *exec.Cmd) error {
			lastCmd = cmd
			return errors.New("bad repository")
		},
		errStr: "unsupported VCS",
	}, {
		name:    "Mercurial",
		input:   []string{"create", "--vcs=hg", "motemen/ghq-hg"},
		want:    []string{"hg", "init"},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghq-hg"),
	}, {
		name:    "Darcs",
		input:   []string{"create", "--vcs=darcs", "motemen/ghq-darcs"},
		want:    []string{"darcs", "init"},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghq-darcs"),
	}, {
		name:    "Bazzar",
		input:   []string{"create", "--vcs=bzr", "motemen/ghq-bzr"},
		want:    []string{"bzr", "init"},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghq-bzr"),
	}, {
		name:    "Fossil",
		input:   []string{"create", "--vcs=fossil", "motemen/ghq-fossil"},
		want:    []string{"fossil", "open", fossilRepoName},
		wantDir: filepath.Join(tmpd, "github.com/motemen/ghq-fossil"),
	}, {
		name:   "unsupported VCS",
		input:  []string{"create", "--vcs=svn", "motemen/ghq-svn"},
		errStr: "unsupported VCS",
	}, {
		name:  "not permitted",
		input: []string{"create", "motemen/ghq-notpermitted"},
		setup: func() func() {
			f := filepath.Join(tmpd, "github.com/motemen/ghq-notpermitted")
			os.MkdirAll(f, 0)
			return func() {
				os.Chmod(f, 0755)
			}
		},
		errStr:    "permission denied",
		skipOnWin: true,
	}, {
		name:  "not empty",
		input: []string{"create", "motemen/ghq-notempty"},
		setup: func() func() {
			f := filepath.Join(tmpd, "github.com/motemen/ghq-notempty", "dummy")
			os.MkdirAll(f, 0755)
			return func() {}
		},
		errStr: "already exists and not empty",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipOnWin && runtime.GOOS == "windows" {
				t.SkipNow()
			}
			lastCmd = nil
			if tc.setup != nil {
				teardown := tc.setup()
				defer teardown()
			}

			cmdutil.CommandRunner = commandRunner
			if tc.cmdRun != nil {
				cmdutil.CommandRunner = tc.cmdRun
			}

			var err error
			out, _, _ := capture(func() {
				err = newApp().Run(append([]string{""}, tc.input...))
			})
			out = strings.TrimSpace(out)

			if tc.errStr == "" {
				if err != nil {
					t.Errorf("error should be nil, but: %s", err)
					return
				}
			} else {
				if err == nil {
					t.Errorf("err should not be nil")
					return
				}
				if e, g := tc.errStr, err.Error(); !strings.Contains(g, e) {
					t.Errorf("err.Error() should contains %q, but not: %q", e, g)
				}
			}

			if len(tc.want) > 0 {
				if !reflect.DeepEqual(lastCmd.Args, tc.want) {
					t.Errorf("cmd.Args = %v, want: %v", lastCmd.Args, tc.want)
				}

				if lastCmd.Dir != tc.wantDir {
					t.Errorf("cmd.Dir = %q, want: %q", lastCmd.Dir, tc.wantDir)
				}
			}

			if tc.errStr == "" {
				if out != tc.wantDir {
					t.Errorf("cmd.Dir = %q, want: %q", out, tc.wantDir)
				}
			} else {
				if out != "" {
					t.Errorf("output should be empty but: %s", out)
				}
			}
		})
	}
}
