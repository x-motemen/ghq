package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/x-motemen/ghq/cmdutil"
)

var (
	remoteDummyURL = mustParseURL("https://example.com/git/repo")
	p4DummyURL     = mustParseURL("p4://alice@p4.example.com:1666/depot/proj/main")
	dummySvnInfo   = []byte(`Path: trunk
URL: https://svn.apache.org/repos/asf/subversion/trunk
Relative URL: ^/subversion/trunk
Repository Root: https://svn.apache.org/repos/asf
Repository UUID: 13f79535-47bb-0310-9956-ffa450edef68
Revision: 1872085
Node Kind: directory
Last Changed Author: julianfoad
Last Changed Rev: 1872031
Last Changed Date: 2019-08-16 15:16:45 +0900 (Fri, 16 Aug 2019)
`)
)

func TestVCSBackend(t *testing.T) {
	tempDir := newTempDir(t)
	localDir := filepath.Join(tempDir, "repo")
	var _commands []*exec.Cmd
	lastCommand := func() *exec.Cmd { return _commands[len(_commands)-1] }
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		_commands = append(_commands, cmd)
		if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
			return fmt.Errorf("[test] failed to svn info")
		}
		return nil
	}
	defer func(orig func() (string, error)) { p4Hostname = orig }(p4Hostname)
	p4Hostname = func() (string, error) { return "myhost", nil }
	p4Dir := filepath.Join(tempDir, "p4repo")
	p4Client := p4ClientName("alice", "myhost", "p4.example.com:1666", "//depot/proj/main", p4Dir)

	testCases := []struct {
		name   string
		f      func() error
		expect []string
		dir    string
	}{{
		name: "[git] clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"git", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] shallow clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
				silent:  true,
			})
		},
		expect: []string{"git", "clone", "--depth", "1", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] clone specific branch",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"git", "clone", "--branch", "hello", "--single-branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] update",
		f: func() error {
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "pull", "--ff-only", "--no-recurse-submodules"},
		dir:    localDir,
	}, {
		name: "[git] fetch",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"git", "rev-parse", "@{upstream}"}) {
					return fmt.Errorf("[test] failed to git rev-parse @{upstream}")
				}
				return nil
			}
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "fetch", "--no-recurse-submodules"},
		dir:    localDir,
	}, {
		name: "[git] recursive",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:       remoteDummyURL,
				dir:       localDir,
				recursive: true,
			})
		},
		expect: []string{"git", "clone", "--recursive", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] update recursive",
		f: func() error {
			return GitBackend.Update(&vcsGetOption{
				dir:       localDir,
				recursive: true,
			})
		},
		expect: []string{"git", "submodule", "foreach", "--recursive",
			"git submodule sync && git -c fetch.recurseSubmodules=false -c submodule.recurse=false submodule update --init"},
		dir: localDir,
	}, {
		name: "[git] bare clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				bare:   true,
				silent: true,
			})
		},
		expect: []string{"git", "clone", "--bare", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] (partial) blobless clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				partial: "blobless",
			})
		},
		expect: []string{"git", "clone", "--filter=blob:none", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] (partial) treeless clone",
		f: func() error {
			return GitBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				partial: "treeless",
			})
		},
		expect: []string{"git", "clone", "--filter=tree:0", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] switch git-svn on update",
		f: func() error {
			err := os.MkdirAll(filepath.Join(localDir, ".git", "svn"), 0755)
			if err != nil {
				return err
			}
			defer os.RemoveAll(filepath.Join(localDir, ".git"))
			return GitBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "rebase"},
		dir:    localDir,
	}, {
		name: "[svn] checkout",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] checkout shallow",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"svn", "checkout", "--depth", "immediates", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] checkout specific branch",
		f: func() error {
			return SubversionBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String() + "/branches/hello", localDir},
	}, {
		name: "[svn] checkout with filling trunk",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			return SubversionBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String() + "/trunk", localDir},
	}, {
		name: "[svn] update",
		f: func() error {
			return SubversionBackend.Update(&vcsGetOption{
				dir:    localDir,
				silent: true,
			})
		},
		expect: []string{"svn", "update"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone",
		f: func() error {
			return GitsvnBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git-svn] update",
		f: func() error {
			return GitsvnBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"git", "svn", "rebase"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone shallow",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(cmd.Args, []string{"svn", "info", "https://example.com/git/repo/trunk"}) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			return GitsvnBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"git", "svn", "clone", "-s", "-r1872031:HEAD", remoteDummyURL.String(), localDir},
	}, {
		name: "[git-svn] clone specific branch",
		f: func() error {
			return GitsvnBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String() + "/branches/hello", localDir},
	}, {
		name: "[git-svn] clone specific branch from tagged URL with shallow",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if reflect.DeepEqual(
					cmd.Args, []string{"svn", "info", "https://example.com/git/repo/branches/develop"},
				) {
					cmd.Stdout.Write(dummySvnInfo)
				}
				return nil
			}
			copied := *remoteDummyURL
			copied.Path += "/tags/v9.9.9"
			return GitsvnBackend.Clone(&vcsGetOption{
				url:     &copied,
				dir:     localDir,
				branch:  "develop",
				shallow: true,
			})
		},
		expect: []string{
			"git", "svn", "clone", "-r1872031:HEAD", remoteDummyURL.String() + "/branches/develop", localDir},
	}, {
		name: "[hg] clone",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] update",
		f: func() error {
			return MercurialBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"hg", "pull", "--update"},
		dir:    localDir,
	}, {
		name: "[hg] clone shallow",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] clone specific branch",
		f: func() error {
			return MercurialBackend.Clone(&vcsGetOption{
				url:    remoteDummyURL,
				dir:    localDir,
				branch: "hello",
			})
		},
		expect: []string{"hg", "clone", "--branch", "hello", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone",
		f: func() error {
			return DarcsBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"darcs", "get", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone shallow",
		f: func() error {
			return DarcsBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"darcs", "get", "--lazy", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] update",
		f: func() error {
			return DarcsBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"darcs", "pull"},
		dir:    localDir,
	}, {
		name: "[pijul] clone",
		f: func() error {
			return PijulBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"pijul", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[pijul] update",
		f: func() error {
			return PijulBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"pijul", "pull"},
		dir:    localDir,
	}, {
		name: "[bzr] clone",
		f: func() error {
			return BazaarBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[bzr] update",
		f: func() error {
			return BazaarBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"bzr", "pull", "--overwrite"},
		dir:    localDir,
	}, {
		name: "[bzr] clone shallow",
		f: func() error {
			return BazaarBackend.Clone(&vcsGetOption{
				url:     remoteDummyURL,
				dir:     localDir,
				shallow: true,
			})
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[fossil] clone",
		f: func() error {
			return FossilBackend.Clone(&vcsGetOption{
				url: remoteDummyURL,
				dir: localDir,
			})
		},
		expect: []string{"fossil", "open", fossilRepoName},
		dir:    localDir,
	}, {
		name: "[fossil] update",
		f: func() error {
			return FossilBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"fossil", "update"},
		dir:    localDir,
	}, {
		name: "[p4] clone",
		f: func() error {
			return PerforceBackend.Clone(&vcsGetOption{
				url: p4DummyURL,
				dir: p4Dir,
			})
		},
		expect: []string{"p4", "-p", "p4.example.com:1666", "-u", "alice", "-c", p4Client, "sync"},
		dir:    p4Dir,
	}, {
		name: "[p4] update",
		f: func() error {
			defer func(orig func(cmd *exec.Cmd) error) {
				cmdutil.CommandRunner = orig
			}(cmdutil.CommandRunner)
			cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
				_commands = append(_commands, cmd)
				if slices.Contains(cmd.Args, "-o") {
					fmt.Fprint(cmd.Stdout, p4TestSpec("ws", localDir, true, "//depot/proj/... //ws/..."))
				}
				return nil
			}
			if err := writeP4Config(localDir, &p4Config{port: "p4.example.com:1666", client: "ws"}); err != nil {
				return err
			}
			defer os.Remove(filepath.Join(localDir, p4ConfigName))
			return PerforceBackend.Update(&vcsGetOption{
				dir: localDir,
			})
		},
		expect: []string{"p4", "-p", "p4.example.com:1666", "-c", "ws", "sync"},
		dir:    localDir,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.f(); err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}
			c := lastCommand()
			if !reflect.DeepEqual(c.Args, tc.expect) {
				t.Errorf("\ngot:    %+v\nexpect: %+v", c.Args, tc.expect)
			}
			if c.Dir != tc.dir {
				t.Errorf("got: %s, expect: %s", c.Dir, tc.dir)
			}
		})
	}
}

func TestGitUpdateSubmoduleURLs(t *testing.T) {
	t.Setenv("GIT_ALLOW_PROTOCOL", "file")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	// TestMain redirects git config; real repositories need their own config.
	t.Setenv("GIT_CONFIG", os.Getenv("GIT_CONFIG"))
	if err := os.Unsetenv("GIT_CONFIG"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name      string
		nested    bool
		recursive bool
		fetchMode string
	}{
		{name: "direct", recursive: true, fetchMode: "on-demand"},
		{name: "nested", nested: true, recursive: true, fetchMode: "false"},
		{name: "no-recursive", nested: true, fetchMode: "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newTempDir(t)
			git := func(dir string, args ...string) string {
				t.Helper()
				cmd := exec.Command("git", append([]string{
					"-c", "user.name=test", "-c", "user.email=test@example.com",
					"-c", "commit.gpgsign=false",
				}, args...)...)
				cmd.Dir = dir
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
				return strings.TrimSpace(string(out))
			}
			initRepo := func(name string) string {
				t.Helper()
				dir := filepath.Join(root, name)
				git(root, "init", "-b", "master", dir)
				git(dir, "commit", "--allow-empty", "-m", "initial")
				return dir
			}

			oldURL := filepath.ToSlash(initRepo("old"))
			oldCommit := git(oldURL, "rev-parse", "HEAD")
			newURL := filepath.ToSlash(filepath.Join(root, "new"))
			git(root, "clone", oldURL, newURL)
			git(newURL, "commit", "--allow-empty", "-m", "new-only commit")
			newCommit := git(newURL, "rev-parse", "HEAD")

			parent := initRepo("parent")
			git(parent, "submodule", "add", oldURL, "child")
			git(parent, "commit", "-am", "add child")
			source := parent
			if tc.nested {
				source = initRepo("outer")
				git(source, "submodule", "add", parent, "middle")
				git(source, "commit", "-am", "add middle")
			}
			local := filepath.Join(root, "local")
			git(root, "clone", "--recursive", source, local)
			git(local, "config", "--local", "fetch.recurseSubmodules", tc.fetchMode)
			if !tc.recursive {
				git(local, "config", "--local", "submodule.recurse", "true")
			}

			git(filepath.Join(parent, "child"), "fetch", newURL, "master")
			git(filepath.Join(parent, "child"), "checkout", "FETCH_HEAD")
			git(parent, "config", "-f", ".gitmodules", "submodule.child.url", newURL)
			git(parent, "commit", "-am", "change child URL and revision")
			if tc.nested {
				git(filepath.Join(source, "middle"), "fetch", "--no-recurse-submodules", parent, "master")
				git(filepath.Join(source, "middle"), "checkout", "FETCH_HEAD")
				git(source, "commit", "-am", "advance middle")
			}

			if err := GitBackend.Update(&vcsGetOption{dir: local, recursive: tc.recursive}); err != nil {
				t.Fatal(err)
			}
			if got, want := git(local, "rev-parse", "HEAD"), git(source, "rev-parse", "HEAD"); got != want {
				t.Errorf("parent HEAD = %s, want %s", got, want)
			}
			childParent := local
			if tc.nested {
				childParent = filepath.Join(local, "middle")
			}
			wantCommit, wantURL := newCommit, newURL
			if !tc.recursive {
				wantCommit, wantURL = oldCommit, oldURL
			}
			if got := git(filepath.Join(childParent, "child"), "rev-parse", "HEAD"); got != wantCommit {
				t.Errorf("child HEAD = %s, want %s", got, wantCommit)
			}
			if got := git(childParent, "config", "--local", "--get", "submodule.child.url"); got != wantURL {
				t.Errorf("child URL = %s, want %s", got, wantURL)
			}
		})
	}
}

func TestCvsDummyBackend(t *testing.T) {
	tempDir := newTempDir(t)
	localDir := filepath.Join(tempDir, "repo")

	if err := cvsDummyBackend.Clone(&vcsGetOption{
		url: remoteDummyURL,
		dir: localDir,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := cvsDummyBackend.Clone(&vcsGetOption{
		url:     remoteDummyURL,
		dir:     localDir,
		shallow: true,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := cvsDummyBackend.Update(&vcsGetOption{
		dir: localDir,
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}
}

func TestBranchOptionIgnoredErrors(t *testing.T) {
	tempDir := newTempDir(t)
	localDir := filepath.Join(tempDir, "repo")

	if err := DarcsBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := FossilBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}

	if err := BazaarBackend.Clone(&vcsGetOption{
		url:    remoteDummyURL,
		dir:    localDir,
		branch: "hello",
	}); err == nil {
		t.Error("error should be occurred, but nil")
	}
}

type p4Call struct {
	args  []string
	dir   string
	stdin string
}

// fakeP4 replaces the command runner with one that records p4 invocations
// and lets respond write output or fail per subcommand ("client -o NAME",
// "client -i", "sync", ...).
func fakeP4(t *testing.T, respond func(sub string, cmd *exec.Cmd) error) *[]p4Call {
	t.Helper()
	origRunner, origHostname := cmdutil.CommandRunner, p4Hostname
	t.Cleanup(func() { cmdutil.CommandRunner, p4Hostname = origRunner, origHostname })
	p4Hostname = func() (string, error) { return "myhost", nil }
	t.Setenv("P4USER", "alice")

	calls := &[]p4Call{}
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		c := p4Call{args: cmd.Args, dir: cmd.Dir}
		if cmd.Stdin != nil {
			b, err := io.ReadAll(cmd.Stdin)
			if err != nil {
				return err
			}
			c.stdin = string(b)
			cmd.Stdin = strings.NewReader(c.stdin)
		}
		*calls = append(*calls, c)
		if cmd.Args[0] != "p4" {
			t.Errorf("unexpected command: %v", cmd.Args)
			return errors.New("unexpected command")
		}
		args := cmd.Args[1:]
		for len(args) >= 2 && (args[0] == "-p" || args[0] == "-u" || args[0] == "-c") {
			args = args[2:]
		}
		if respond == nil {
			return nil
		}
		return respond(strings.Join(args, " "), cmd)
	}
	return calls
}

// p4TestSpec returns a client spec as printed by "p4 client -o".
func p4TestSpec(client, root string, exists bool, view ...string) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "# A Perforce Client Specification.\n#\n#  Root:        The base directory of the client workspace.\n\nClient:\t%s\n\n", client)
	if exists {
		b.WriteString("Update:\t2026/01/02 03:04:05\n\nAccess:\t2026/01/02 03:04:05\n\n")
	}
	fmt.Fprintf(b, "Owner:\talice\n\nHost:\tmyhost\n\nDescription:\n\tCreated by alice.\n\nRoot:\t%s\n\n", root)
	b.WriteString("Options:\tnoallwrite noclobber nocompress unlocked nomodtime normdir noaltsync\n\nView:\n")
	for _, v := range view {
		fmt.Fprintf(b, "\t%s\n", v)
	}
	return b.String()
}

func readFileString(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPerforceBackend_Clone(t *testing.T) {
	const port = "p4.example.com:1666"
	newDir := func(t *testing.T) string {
		return filepath.Join(newTempDir(t), "p4.example.com", "depot", "proj", "main")
	}

	t.Run("new client", func(t *testing.T) {
		dir := newDir(t)
		client := p4ClientName("alice", "myhost", port, "//depot/proj/main", dir)
		calls := fakeP4(t, func(sub string, cmd *exec.Cmd) error {
			if sub == "client -o "+client {
				fmt.Fprint(cmd.Stdout, p4TestSpec(client, "/home/alice", false, "//depot/... //"+client+"/..."))
			}
			return nil
		})
		err := PerforceBackend.Clone(&vcsGetOption{
			url: mustParseURL("p4://p4.example.com/depot/proj/main/"),
			dir: dir,
			// ignored by p4
			shallow: true, recursive: true, partial: "blobless", silent: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		global := []string{"p4", "-p", port, "-c", client}
		expect := []p4Call{
			{args: append(global[:5:5], "client", "-o", client)},
			{args: append(global[:5:5], "client", "-i")},
			{args: append(global[:5:5], "sync"), dir: dir},
		}
		if len(*calls) != len(expect) {
			t.Fatalf("got %d commands, expect %d: %+v", len(*calls), len(expect), *calls)
		}
		for i, c := range *calls {
			if !reflect.DeepEqual(c.args, expect[i].args) || c.dir != expect[i].dir {
				t.Errorf("command %d:\ngot:    %v (dir %q)\nexpect: %v (dir %q)", i, c.args, c.dir, expect[i].args, expect[i].dir)
			}
		}
		spec := (*calls)[1].stdin
		for _, want := range []string{
			"\nOwner:\talice\n",
			"\nRoot:\t" + dir + "\n",
			"\nView:\n\t\"//depot/proj/main/...\" \"//" + client + "/...\"\n\t\"-//depot/proj/main/.p4config\" \"//" + client + "/.p4config\"\n",
		} {
			if !strings.Contains(spec, want) {
				t.Errorf("client spec should contain %q, but:\n%s", want, spec)
			}
		}
		if strings.Contains(spec, "/home/alice") || strings.Contains(spec, "//depot/... ") {
			t.Errorf("client spec should not keep the default root and view:\n%s", spec)
		}
		if got, expect := readFileString(t, filepath.Join(dir, p4ConfigName)), "P4PORT="+port+"\nP4CLIENT="+client+"\n"; got != expect {
			t.Errorf(".p4config:\ngot:    %q\nexpect: %q", got, expect)
		}
	})

	t.Run("ssl with user", func(t *testing.T) {
		dir := newDir(t)
		const sslPort = "ssl:p4.example.com:1667"
		client := p4ClientName("bob", "myhost", sslPort, "//depot/proj/main", dir)
		calls := fakeP4(t, nil)
		err := PerforceBackend.Clone(&vcsGetOption{
			url: mustParseURL("p4+ssl://bob@p4.example.com:1667/depot/proj/main"),
			dir: dir,
		})
		if err != nil {
			t.Fatal(err)
		}
		expect := []string{"p4", "-p", sslPort, "-u", "bob", "-c", client, "sync"}
		if got := (*calls)[len(*calls)-1].args; !reflect.DeepEqual(got, expect) {
			t.Errorf("\ngot:    %v\nexpect: %v", got, expect)
		}
		if got, expect := readFileString(t, filepath.Join(dir, p4ConfigName)), "P4PORT="+sslPort+"\nP4USER=bob\nP4CLIENT="+client+"\n"; got != expect {
			t.Errorf(".p4config:\ngot:    %q\nexpect: %q", got, expect)
		}
	})

	t.Run("reuse client left by ghq rm", func(t *testing.T) {
		dir := newDir(t)
		client := p4ClientName("alice", "myhost", port, "//depot/proj/main", dir)
		calls := fakeP4(t, func(sub string, cmd *exec.Cmd) error {
			if sub == "client -o "+client {
				fmt.Fprint(cmd.Stdout, p4TestSpec(client, dir, true, "//depot/proj/main/... //"+client+"/..."))
			}
			return nil
		})
		if err := PerforceBackend.Clone(&vcsGetOption{url: mustParseURL("p4://p4.example.com/depot/proj/main"), dir: dir}); err != nil {
			t.Fatal(err)
		}
		expect := []string{"p4", "-p", port, "-c", client, "sync", "-f"}
		if got := (*calls)[len(*calls)-1].args; !reflect.DeepEqual(got, expect) {
			t.Errorf("\ngot:    %v\nexpect: %v", got, expect)
		}
	})

	failures := []struct {
		name    string
		url     string
		branch  string
		bare    bool
		mkdir   bool
		respond func(client, dir string) func(sub string, cmd *exec.Cmd) error
		errStr  string
		calls   int
	}{{
		name: "client of another workspace",
		url:  "p4://p4.example.com/depot/proj/main",
		respond: func(client, dir string) func(string, *exec.Cmd) error {
			return func(sub string, cmd *exec.Cmd) error {
				if strings.HasPrefix(sub, "client -o") {
					fmt.Fprint(cmd.Stdout, p4TestSpec(client, "/somewhere/else", true, "//depot/... //"+client+"/..."))
				}
				return nil
			}
		},
		errStr: "already exists with root",
		calls:  1,
	}, {
		name: "not logged in",
		url:  "p4://p4.example.com/depot/proj/main",
		respond: func(client, dir string) func(string, *exec.Cmd) error {
			return func(sub string, cmd *exec.Cmd) error {
				fmt.Fprintln(cmd.Stderr, "Perforce password (P4PASSWD) invalid or unset.")
				return errors.New("exit status 1")
			}
		},
		errStr: "Perforce password (P4PASSWD) invalid or unset.",
		calls:  1,
	}, {
		name: "client spec rejected",
		url:  "p4://p4.example.com/nodepot/proj",
		respond: func(client, dir string) func(string, *exec.Cmd) error {
			return func(sub string, cmd *exec.Cmd) error {
				if sub == "client -i" {
					fmt.Fprintln(cmd.Stderr, "Error in client specification.")
					return errors.New("exit status 1")
				}
				return nil
			}
		},
		errStr: "Error in client specification.",
		calls:  2,
	}, {
		name: "sync fails",
		url:  "p4://p4.example.com/depot/proj/main",
		respond: func(client, dir string) func(string, *exec.Cmd) error {
			return func(sub string, cmd *exec.Cmd) error {
				if sub == "sync" {
					return errors.New("exit status 1")
				}
				return nil
			}
		},
		errStr: "exit status 1",
		calls:  3,
	}, {
		name:   "existing directory",
		url:    "p4://p4.example.com/depot/proj/main",
		mkdir:  true,
		errStr: "already exists",
	}, {
		name:   "branch",
		url:    "p4://p4.example.com/depot/proj/main",
		branch: "dev",
		errStr: "--branch option",
	}, {
		name:   "bare",
		url:    "p4://p4.example.com/depot/proj/main",
		bare:   true,
		errStr: "bare checkouts are unavailable",
	}, {
		name:   "no host",
		url:    "p4:///depot/proj",
		errStr: "missing p4 server host",
	}, {
		name:   "no depot path",
		url:    "p4://p4.example.com/",
		errStr: "missing depot path",
	}, {
		name:   "password",
		url:    "p4://alice:secret@p4.example.com/depot/proj",
		errStr: "passwords in p4 URLs are not supported",
	}, {
		name:   "newline in user",
		url:    "p4://al%0Aice@p4.example.com/depot/proj",
		errStr: "invalid p4 user",
	}, {
		name:   "wildcard",
		url:    "p4://p4.example.com/depot/proj/...",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "asterisk",
		url:    "p4://p4.example.com/depot/*/main",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "revision",
		url:    "p4://p4.example.com/depot/proj%23head",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "label",
		url:    "p4://p4.example.com/depot/proj%40label",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "percent escape",
		url:    "p4://p4.example.com/depot/proj%2540",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "quote",
		url:    "p4://p4.example.com/depot/pr%22oj",
		errStr: "wildcards and revision specifiers",
	}, {
		name:   "spec injection",
		url:    "p4://p4.example.com/depot/proj%0AOptions:%09allwrite",
		errStr: "control characters",
	}, {
		name:   "empty component",
		url:    "p4://p4.example.com/depot//proj",
		errStr: "empty or dot path components",
	}, {
		name:   "parent component",
		url:    "p4://p4.example.com/depot/../proj",
		errStr: "empty or dot path components",
	}}
	for _, tc := range failures {
		t.Run(tc.name, func(t *testing.T) {
			dir := newDir(t)
			if tc.mkdir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}
			u := mustParseURL(tc.url)
			var respond func(string, *exec.Cmd) error
			if tc.respond != nil {
				cfg, depotPath, err := p4ConfigFromURL(u)
				if err != nil {
					t.Fatal(err)
				}
				respond = tc.respond(p4ClientName("alice", "myhost", cfg.port, depotPath, dir), dir)
			}
			calls := fakeP4(t, respond)
			err := PerforceBackend.Clone(&vcsGetOption{url: u, dir: dir, branch: tc.branch, bare: tc.bare})
			if err == nil || !strings.Contains(err.Error(), tc.errStr) {
				t.Errorf("error should contain %q, but: %v", tc.errStr, err)
			}
			if len(*calls) != tc.calls {
				t.Errorf("got %d commands, expect %d: %+v", len(*calls), tc.calls, *calls)
			}
			_, statErr := os.Stat(dir)
			if tc.mkdir {
				if statErr != nil {
					t.Errorf("existing directory should be kept: %v", statErr)
				}
			} else if !os.IsNotExist(statErr) {
				t.Errorf("directory should not be left behind: %v", statErr)
			}
		})
	}
}

func TestPerforceBackend_Update(t *testing.T) {
	t.Run("with user", func(t *testing.T) {
		dir := newTempDir(t)
		config := "# comment=ignored\r\nP4PORT = ssl:p4.example.com:1666\r\nP4USER=bob\r\nP4CLIENT=bob_ws\r\nP4CHARSET=utf8\r\n"
		if err := os.WriteFile(filepath.Join(dir, p4ConfigName), []byte(config), 0644); err != nil {
			t.Fatal(err)
		}
		srv := newP4FakeServer()
		srv.clients["bob_ws"] = p4TestSpec("bob_ws", dir, true, "//depot/proj/... //bob_ws/...")
		calls := fakeP4(t, srv.respond)
		if err := PerforceBackend.Update(&vcsGetOption{dir: dir, silent: true, url: mustParseURL("p4+ssl://bob@p4.example.com/depot/proj")}); err != nil {
			t.Fatal(err)
		}
		global := []string{"p4", "-p", "ssl:p4.example.com:1666", "-u", "bob", "-c", "bob_ws"}
		expect := []p4Call{
			{args: append(global[:7:7], "client", "-o", "bob_ws")},
			{args: append(global[:7:7], "sync"), dir: dir},
		}
		if !reflect.DeepEqual(*calls, expect) {
			t.Errorf("\ngot:    %+v\nexpect: %+v", *calls, expect)
		}
	})

	for name, config := range map[string]*string{
		"no .p4config": nil,
		"no client":    ptr("P4PORT=p4.example.com:1666\n"),
		"no port":      ptr("P4CLIENT=ws\n"),
	} {
		t.Run(name, func(t *testing.T) {
			dir := newTempDir(t)
			if config != nil {
				if err := os.WriteFile(filepath.Join(dir, p4ConfigName), []byte(*config), 0644); err != nil {
					t.Fatal(err)
				}
			}
			calls := fakeP4(t, nil)
			if err := PerforceBackend.Update(&vcsGetOption{dir: dir}); err == nil {
				t.Error("error should be occurred, but nil")
			}
			if len(*calls) != 0 {
				t.Errorf("p4 should not be run: %+v", *calls)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func TestPerforceBackend_RemoteURL(t *testing.T) {
	testCases := []struct {
		name   string
		config string
		spec   func(client, root string) string
		expect string
		errStr string
	}{{
		name:   "tcp",
		config: "P4PORT=p4.example.com:1666\nP4CLIENT=ws\n",
		spec: func(c, root string) string {
			return p4TestSpec(c, root, true, "//depot/proj/main/... //ws/...", "-//depot/proj/main/.p4config //ws/.p4config")
		},
		expect: "p4://p4.example.com:1666/depot/proj/main",
	}, {
		name:   "ssl with user",
		config: "P4PORT=ssl:p4.example.com:1667\nP4USER=bob\nP4CLIENT=ws\n",
		spec: func(c, root string) string {
			return p4TestSpec(c, root, true, "-//depot/x/... //ws/x/...", "//depot/proj/... //ws/...")
		},
		expect: "p4+ssl://bob@p4.example.com:1667/depot/proj",
	}, {
		name:   "quoted view",
		config: "P4PORT=1666\nP4CLIENT=ws\n",
		spec: func(c, root string) string {
			return p4TestSpec(c, root, true, `"//depot/my proj/..." "//ws/..."`)
		},
		expect: "p4://localhost:1666/depot/my%20proj",
	}, {
		name:   "client deleted on server",
		config: "P4PORT=p4.example.com:1666\nP4CLIENT=ws\n",
		spec: func(c, root string) string {
			return p4TestSpec(c, root, false, "//depot/... //ws/...")
		},
		errStr: "does not exist",
	}, {
		name:   "no depot path in view",
		config: "P4PORT=p4.example.com:1666\nP4CLIENT=ws\n",
		spec: func(c, root string) string {
			return p4TestSpec(c, root, true, "//depot/file.txt //ws/file.txt")
		},
		errStr: "no depot path found",
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := newTempDir(t)
			if err := os.WriteFile(filepath.Join(dir, p4ConfigName), []byte(tc.config), 0644); err != nil {
				t.Fatal(err)
			}
			calls := fakeP4(t, func(sub string, cmd *exec.Cmd) error {
				if sub == "client -o ws" {
					fmt.Fprint(cmd.Stdout, tc.spec("ws", dir))
				}
				return nil
			})
			got, err := PerforceBackend.RemoteURL(dir)
			if tc.errStr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errStr) {
					t.Errorf("error should contain %q, but: %v", tc.errStr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.expect {
				t.Errorf("got: %s, expect: %s", got, tc.expect)
			}
			if len(*calls) != 1 || !slices.Contains((*calls)[0].args, "-c") {
				t.Errorf("unexpected commands: %+v", *calls)
			}
		})
	}
}

func TestP4UpdateClientRoot(t *testing.T) {
	dir := newTempDir(t)
	if err := writeP4Config(dir, &p4Config{port: "p4.example.com:1666", client: "ws"}); err != nil {
		t.Fatal(err)
	}
	calls := fakeP4(t, func(sub string, cmd *exec.Cmd) error {
		if sub == "client -o ws" {
			fmt.Fprint(cmd.Stdout, p4TestSpec("ws", "/old/root", true, "//depot/proj/... //ws/..."))
		}
		return nil
	})
	if err := p4UpdateClientRoot("/old/root", dir); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("unexpected commands: %+v", *calls)
	}
	spec := (*calls)[1].stdin
	if !strings.Contains(spec, "\nRoot:\t"+dir+"\n") || strings.Contains(spec, "/old/root") ||
		!strings.Contains(spec, "\n\t//depot/proj/... //ws/...\n") {
		t.Errorf("unexpected client spec:\n%s", spec)
	}
}

func TestP4ConfigFromURL(t *testing.T) {
	testCases := []struct {
		url, port, user, depot string
	}{
		{"p4://p4.example.com/depot/proj/main", "p4.example.com:1666", "", "//depot/proj/main"},
		{"p4://p4.example.com:1777/depot/proj/main/", "p4.example.com:1777", "", "//depot/proj/main"},
		{"p4://alice@p4.example.com/depot/my%20proj", "p4.example.com:1666", "alice", "//depot/my proj"},
		{"p4+ssl://p4.example.com/depot", "ssl:p4.example.com:1666", "", "//depot"},
		{"p4://[::1]:1666/depot/proj", "tcp6:[::1]:1666", "", "//depot/proj"},
		{"p4+ssl://[::1]/depot/proj", "ssl6:[::1]:1666", "", "//depot/proj"},
		// ghq.<url>.vcs = p4 or --vcs=p4 without a p4:// URL
		{"https://p4.example.com/depot/proj", "p4.example.com:1666", "", "//depot/proj"},
	}
	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			cfg, depot, err := p4ConfigFromURL(mustParseURL(tc.url))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.port != tc.port || cfg.user != tc.user || depot != tc.depot {
				t.Errorf("got: (%q, %q, %q), expect: (%q, %q, %q)", cfg.port, cfg.user, depot, tc.port, tc.user, tc.depot)
			}
			// RemoteURL must rebuild an equivalent URL
			u, err := p4URLFromConfig(cfg, depot)
			if err != nil {
				t.Fatal(err)
			}
			cfg2, depot2, err := p4ConfigFromURL(u)
			if err != nil {
				t.Fatal(err)
			}
			if *cfg2 != *cfg || depot2 != depot {
				t.Errorf("round trip via %s: got (%+v, %q), expect (%+v, %q)", u, cfg2, depot2, cfg, depot)
			}
		})
	}
}

func TestP4URLFromConfig(t *testing.T) {
	testCases := []struct {
		port, expect string
	}{
		{"1666", "p4://localhost:1666/depot/proj"},
		{"p4.example.com:1666", "p4://p4.example.com:1666/depot/proj"},
		{"tcp:p4.example.com:1666", "p4://p4.example.com:1666/depot/proj"},
		{"tcp4:p4.example.com:1666", "p4://p4.example.com:1666/depot/proj"},
		{"ssl:p4.example.com:1666", "p4+ssl://p4.example.com:1666/depot/proj"},
		{"ssl:1666", "p4+ssl://localhost:1666/depot/proj"},
		{"tcp6:[::1]:1666", "p4://[::1]:1666/depot/proj"},
		{"perforce:1666", "p4://perforce:1666/depot/proj"},
		{"rsh:p4d -r /p4root -i", ""},
		{"p4.example.com", ""},
		{"p4.example.com:", ""},
	}
	for _, tc := range testCases {
		t.Run(tc.port, func(t *testing.T) {
			u, err := p4URLFromConfig(&p4Config{port: tc.port, client: "ws"}, "//depot/proj")
			if tc.expect == "" {
				if err == nil {
					t.Errorf("error should be occurred, but got %s", u)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if u.String() != tc.expect {
				t.Errorf("got: %s, expect: %s", u, tc.expect)
			}
		})
	}
}

func TestP4ClientName(t *testing.T) {
	const port = "p4.example.com:1666"
	name := p4ClientName("alice", "my.host", port, "//depot/proj/main", "/ghq/p4.example.com/depot/proj/main")
	if !regexp.MustCompile(`^alice_my_host_depot_proj_main_[0-9a-f]{8}$`).MatchString(name) {
		t.Errorf("unexpected client name: %s", name)
	}
	if again := p4ClientName("alice", "my.host", port, "//depot/proj/main", "/ghq/p4.example.com/depot/proj/main"); again != name {
		t.Errorf("client name should be deterministic: %s != %s", again, name)
	}

	// Names that sanitize to the same slug must still differ.
	seen := map[string]bool{}
	for _, in := range [][2]string{
		{"//depot/a_b", "/ghq/h/depot/a_b"},
		{"//depot/a/b", "/ghq/h/depot/a/b"},
		{"//depot/a b", "/ghq/h/depot/a b"},
		{"//depot/a/b", "/other-root/h/depot/a/b"},
	} {
		n := p4ClientName("alice", "my.host", port, in[0], in[1])
		if seen[n] {
			t.Errorf("client name collision: %s", n)
		}
		seen[n] = true
	}
	if p4ClientName("alice", "h", port, "//depot/a", "/r") == p4ClientName("alice", "h", "ssl:"+port, "//depot/a", "/r") {
		t.Error("client names should differ between servers")
	}

	valid := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
	for _, in := range [][3]string{
		{"-u x", "host;rm -rf", "//depot/..."},
		{`DOMAIN\user`, "HOST", "//depot/@#%*"},
		{"ユーザー", "ホスト", "//デポ/パス"},
		{"", "", "//depot/" + strings.Repeat("long/", 100)},
	} {
		n := p4ClientName(in[0], in[1], port, in[2], "/r")
		if !valid.MatchString(n) || len(n) > 140 {
			t.Errorf("invalid client name for %q: %q", in, n)
		}
	}
	if n := p4ClientName("ユーザー", "ホスト", port, "//デポ", "/r"); !strings.HasPrefix(n, "ghq_") {
		t.Errorf("client name without readable parts should start with ghq_: %s", n)
	}
}

func TestP4Spec(t *testing.T) {
	spec := "# Root: comment\r\nClient:\tws\r\n\r\nRoot:\t/old root\r\n\r\nAltRoots:\r\n\t/alt\r\n\r\nView:\r\n\t//depot/a/... //ws/a/...\r\n\t\"//depot/b c/...\" \"//ws/b c/...\"\r\n\r\nLineEnd:\tlocal\r\n"
	if got := p4SpecField(spec, "Root"); !reflect.DeepEqual(got, []string{"/old root"}) {
		t.Errorf("Root: got %q", got)
	}
	if got := p4SpecField(spec, "View"); !reflect.DeepEqual(got, []string{"//depot/a/... //ws/a/...", `"//depot/b c/..." "//ws/b c/..."`}) {
		t.Errorf("View: got %q", got)
	}
	if got := p4SpecField(spec, "Access"); got != nil {
		t.Errorf("Access: got %q", got)
	}
	if got := p4SpecField("Description:\n\nRoot:\t/r\n", "Description"); got == nil || len(got) != 0 {
		t.Errorf("empty Description: got %#v", got)
	}

	edited := p4SetSpecField(spec, "Root", false, "/new root")
	edited = p4SetSpecField(edited, "View", true, "//depot/x/... //ws/...")
	edited = p4SetSpecField(edited, "Stream", false, "//stream/main")
	expect := "# Root: comment\nClient:\tws\n\nRoot:\t/new root\n\nAltRoots:\n\t/alt\n\nView:\n\t//depot/x/... //ws/...\n\nLineEnd:\tlocal\n\n\nStream:\t//stream/main"
	if edited != expect {
		t.Errorf("\ngot:    %q\nexpect: %q", edited, expect)
	}

	for _, tc := range []struct {
		view   []string
		expect string
	}{
		{[]string{"//depot/a/... //ws/..."}, "//depot/a"},
		{[]string{`"-//depot/a/x" "//ws/x"`, `"//depot/b c/..." "//ws/..."`}, "//depot/b c"},
		{[]string{"+//depot/o/... //ws/...", "//depot/c/... //ws/c/..."}, "//depot/c"},
		{[]string{"//depot/... //ws/..."}, "//depot"},
		{[]string{"//... //ws/..."}, ""},
		{nil, ""},
	} {
		got, err := p4ViewDepotPath(tc.view)
		if got != tc.expect || (err == nil) != (tc.expect != "") {
			t.Errorf("p4ViewDepotPath(%q) = %q, %v; expect %q", tc.view, got, err, tc.expect)
		}
	}
}

// p4FakeServer is an in-memory stand-in for a p4 server's client specs.
type p4FakeServer struct {
	clients map[string]string // name -> spec
	fail    map[string]string // subcommand -> error message
}

func newP4FakeServer() *p4FakeServer {
	return &p4FakeServer{clients: map[string]string{}, fail: map[string]string{}}
}

func (s *p4FakeServer) respond(sub string, cmd *exec.Cmd) error {
	if msg, ok := s.fail[sub]; ok {
		fmt.Fprintln(cmd.Stderr, msg)
		return errors.New("exit status 1")
	}
	switch {
	case strings.HasPrefix(sub, "client -o "):
		name := strings.TrimPrefix(sub, "client -o ")
		if spec, ok := s.clients[name]; ok {
			fmt.Fprint(cmd.Stdout, spec)
		} else {
			fmt.Fprint(cmd.Stdout, p4TestSpec(name, "/home/alice", false, "//depot/... //"+name+"/..."))
		}
	case sub == "client -i":
		b, err := io.ReadAll(cmd.Stdin)
		if err != nil {
			return err
		}
		spec := string(b)
		if p4SpecField(spec, "Access") == nil {
			spec = strings.Replace(spec, "\nOwner:", "\nAccess:\t2026/01/02 03:04:05\n\nOwner:", 1)
		}
		s.clients[p4SpecField(spec, "Client")[0]] = spec
	case sub == "sync" || sub == "sync -f":
		return os.WriteFile(filepath.Join(cmd.Dir, "README"), []byte("hello\n"), 0644)
	}
	return nil
}

// Regression tests for the first adversarial review round.
func TestPerforceBackend_Review1(t *testing.T) {
	const port = "p4.example.com:1666"
	writeConfig := func(t *testing.T, dir, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, p4ConfigName), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("rsh P4PORT in .p4config is never passed to p4", func(t *testing.T) {
		dir := newTempDir(t)
		writeConfig(t, dir, "P4PORT=rsh:touch pwned\nP4CLIENT=ws\n")
		calls := fakeP4(t, nil)
		if err := PerforceBackend.Update(&vcsGetOption{dir: dir}); err == nil {
			t.Error("Update: error should be occurred, but nil")
		}
		if _, err := PerforceBackend.RemoteURL(dir); err == nil {
			t.Error("RemoteURL: error should be occurred, but nil")
		}
		if len(*calls) != 0 {
			t.Errorf("p4 should not be run: %+v", *calls)
		}
		if b := findVCSBackend(dir, ""); b != nil {
			t.Errorf("a .p4config with an unsupported P4PORT should not mark a workspace: %v", b)
		}
	})

	t.Run("get -u from another server", func(t *testing.T) {
		dir := newTempDir(t)
		writeConfig(t, dir, "P4PORT="+port+"\nP4CLIENT=ws\n")
		srv := newP4FakeServer()
		srv.clients["ws"] = p4TestSpec("ws", dir, true, "//depot/proj/... //ws/...")
		calls := fakeP4(t, srv.respond)
		for _, u := range []string{
			"p4+ssl://p4.example.com:1667/depot/proj",
			"p4://p4.example.com:1667/depot/proj",
			"p4://bob@p4.example.com:1666/depot/proj",
			"p4://p4.example.com/depot/proj#123",
		} {
			if err := PerforceBackend.Update(&vcsGetOption{dir: dir, url: mustParseURL(u)}); err == nil {
				t.Errorf("%s: error should be occurred, but nil", u)
			}
		}
		for _, c := range *calls {
			if slices.Contains(c.args, "sync") {
				t.Errorf("should not sync: %v", c.args)
			}
		}
		// equivalent spellings of the same server are accepted
		for _, u := range []string{"p4://P4.example.com/depot/proj", "p4://p4.example.com:1666/depot/proj", "https://p4.example.com/depot/proj"} {
			if err := PerforceBackend.Update(&vcsGetOption{dir: dir, url: mustParseURL(u)}); err != nil {
				t.Errorf("%s: %v", u, err)
			}
		}
	})

	t.Run("get -u of a client rooted elsewhere or deleted", func(t *testing.T) {
		dir := newTempDir(t)
		writeConfig(t, dir, "P4PORT="+port+"\nP4CLIENT=ws\n")
		for name, spec := range map[string]string{
			"elsewhere": p4TestSpec("ws", "/elsewhere", true, "//depot/proj/... //ws/..."),
			"deleted":   p4TestSpec("ws", "/home/alice", false, "//depot/... //ws/..."),
		} {
			t.Run(name, func(t *testing.T) {
				srv := newP4FakeServer()
				srv.clients["ws"] = spec
				calls := fakeP4(t, srv.respond)
				if err := PerforceBackend.Update(&vcsGetOption{dir: dir}); err == nil {
					t.Error("error should be occurred, but nil")
				}
				if _, err := PerforceBackend.RemoteURL(dir); err == nil {
					t.Error("RemoteURL: error should be occurred, but nil")
				}
				for _, c := range *calls {
					if slices.Contains(c.args, "sync") {
						t.Errorf("should not sync: %v", c.args)
					}
				}
			})
		}
	})

	t.Run("clone into a directory created concurrently", func(t *testing.T) {
		dir := filepath.Join(newTempDir(t), "p4.example.com", "depot", "proj")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "theirs"), nil, 0644); err != nil {
			t.Fatal(err)
		}
		srv := newP4FakeServer()
		srv.fail["sync"] = "sync failed"
		calls := fakeP4(t, srv.respond)
		if err := PerforceBackend.Clone(&vcsGetOption{url: mustParseURL("p4://p4.example.com/depot/proj"), dir: dir}); err == nil {
			t.Error("error should be occurred, but nil")
		}
		if _, err := os.Stat(filepath.Join(dir, "theirs")); err != nil {
			t.Errorf("files of the other checkout should be kept: %v", err)
		}
		if len(*calls) != 0 {
			t.Errorf("p4 should not be run: %+v", *calls)
		}
	})

	t.Run("revision fragments and queries", func(t *testing.T) {
		for _, u := range []string{"p4://p4.example.com/depot/proj#123", "p4://p4.example.com/depot/proj?rev=1"} {
			dir := filepath.Join(newTempDir(t), "p4.example.com", "depot", "proj")
			calls := fakeP4(t, newP4FakeServer().respond)
			if err := PerforceBackend.Clone(&vcsGetOption{url: mustParseURL(u), dir: dir}); err == nil {
				t.Errorf("%s: error should be occurred, but nil", u)
			}
			if len(*calls) != 0 {
				t.Errorf("%s: p4 should not be run: %+v", u, *calls)
			}
		}
	})

	t.Run("client names of similar hosts and users differ", func(t *testing.T) {
		for _, pair := range [][2][2]string{
			{{"alice", "build.runner"}, {"alice", "build_runner"}},
			{{"alice", "ci-runner-" + strings.Repeat("x", 40) + "-1"}, {"alice", "ci-runner-" + strings.Repeat("x", 40) + "-2"}},
			{{"a.b", "host"}, {"a_b", "host"}},
		} {
			a := p4ClientName(pair[0][0], pair[0][1], port, "//depot/proj", "/ghq/p4.example.com/depot/proj")
			b := p4ClientName(pair[1][0], pair[1][1], port, "//depot/proj", "/ghq/p4.example.com/depot/proj")
			if a == b {
				t.Errorf("client names should differ for %q: %s", pair, a)
			}
		}
	})
}
