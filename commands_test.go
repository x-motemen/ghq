package main

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/urfave/cli/v3"
	"github.com/x-motemen/ghq/cmdutil"
)

func TestVCSFlagRejectsUnknown(t *testing.T) {
	for _, tc := range []struct {
		name, command string
		flags         []string
		existing      bool
		fromStdin     bool
		input         []string
	}{
		{name: "get", command: "get"},
		{name: "clone alias", command: "clone"},
		{name: "parallel", command: "get", flags: []string{"--parallel"}},
		{name: "existing", command: "get", existing: true},
		{name: "update", command: "get", flags: []string{"--update"}, existing: true},
		{name: "stdin", command: "get", fromStdin: true, input: []string{"motemen/ghq-test-repo"}},
		{name: "empty stdin", command: "get", fromStdin: true},
		{name: "parallel empty stdin", command: "get", flags: []string{"--parallel"}, fromStdin: true},
		{name: "create", command: "create"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newTempDir(t)
			t.Setenv(envGhqRoot, root)
			defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
			_localRepositoryRoots = []string{root}
			if tc.existing {
				if err := os.MkdirAll(filepath.Join(root, "github.com/motemen/ghq-test-repo/.git"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			defer func(orig func(*exec.Cmd) error) { cmdutil.CommandRunner = orig }(cmdutil.CommandRunner)
			var commands []string
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				commands = append(commands, cmd.String())
				return nil
			}
			args := append([]string{"ghq", tc.command}, tc.flags...)
			args = append(args, "--vcs=unknown-vcs")
			if !tc.fromStdin {
				args = append(args, "motemen/ghq-test-repo")
			}
			var runErr error
			out, _, err := captureWithInput(tc.input, func() {
				runErr = newApp().Run(context.Background(), args)
			})
			if err != nil {
				t.Fatal(err)
			}
			if runErr == nil || !strings.Contains(runErr.Error(), `unknown VCS "unknown-vcs"`) {
				t.Errorf("expected unknown VCS error identifying the value, got: %v", runErr)
			}
			if len(commands) != 0 {
				t.Errorf("unexpected VCS commands: %v", commands)
			}
			if out != "" {
				t.Errorf("unexpected output: %q", out)
			}
			if !tc.existing {
				entries, err := os.ReadDir(root)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 0 {
					t.Errorf("repository directories created for an unknown VCS: %v", entries)
				}
			}
		})
	}
}

func TestVCSFlagAcceptsKnownValues(t *testing.T) {
	for _, command := range []*cli.Command{commandGet, commandCreate} {
		for _, vcs := range []string{"", "git", "github", "codecommit", "svn", "subversion", "git-svn", "hg", "mercurial", "darcs", "pijul", "fossil", "bzr", "bazaar"} {
			t.Run(command.Name+"/"+vcs, func(t *testing.T) {
				cmd := *command
				called := false
				cmd.Action = func(context.Context, *cli.Command) error {
					called = true
					return nil
				}
				app := newApp()
				app.Commands = []*cli.Command{&cmd}
				if err := app.Run(context.Background(), []string{"ghq", cmd.Name, "--vcs=" + vcs}); err != nil {
					t.Fatal(err)
				}
				if !called {
					t.Error("command action was not called")
				}
			})
		}
	}
}

type _cloneArgs struct {
	remote    *url.URL
	local     string
	shallow   bool
	branch    string
	recursive bool
	bare      bool
	silent    bool
	partial   string
}

type _updateArgs struct {
	local string
	bare  bool
}

func withFakeGitBackend(t *testing.T, block func(*testing.T, string, *_cloneArgs, *_updateArgs)) {
	tmpRoot := newTempDir(t)

	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	_localRepositoryRoots = []string{tmpRoot}

	var cloneArgs _cloneArgs
	var updateArgs _updateArgs

	var originalGitBackend = GitBackend
	tmpBackend := &VCSBackend{
		Clone: func(vg *vcsGetOption) error {
			cloneArgs = _cloneArgs{
				remote:    vg.url,
				local:     filepath.FromSlash(vg.dir),
				shallow:   vg.shallow,
				branch:    vg.branch,
				recursive: vg.recursive,
				bare:      vg.bare,
				silent:    vg.silent,
				partial:   vg.partial,
			}
			return nil
		},
		Update: func(vg *vcsGetOption) error {
			updateArgs = _updateArgs{
				local: vg.dir,
				bare:  vg.bare,
			}
			return nil
		},
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}

	GitBackend = tmpBackend
	vcsContentsMap[".git"] = tmpBackend
	defer func() { GitBackend = originalGitBackend; vcsContentsMap[".git"] = originalGitBackend }()
	block(t, tmpRoot, &cloneArgs, &updateArgs)
}
