package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/urfave/cli/v3"
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
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: false,
		},
		{
			name:  "unknown host with one path component",
			input: []string{"rm", "https://example.com/repo"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "example.com", "repo", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:  "empty directory",
			input: []string{"rm", "motemen/ghqqq"},
			setup: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqqq"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: true,
		},
		{
			name:      "missing directory",
			input:     []string{"rm", "motemen/missing"},
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

			var runErr error
			_, _, err := captureWithInput([]string{"y"}, func() {
				a := newApp()
				args := append([]string{"ghq"}, tc.input...)
				runErr = a.Run(context.Background(), args)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotErr := runErr != nil; gotErr != tc.expectErr {
				t.Fatalf("error = %v, expectErr = %v", runErr, tc.expectErr)
			}
		})
	}
}

func TestRmMultipleRepositories(t *testing.T) {
	root := newTempDir(t)
	setRepositoryPathTestRoot(t, root)
	names := []string{"motemen/ghq-first", "motemen/ghq-second"}
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(root, "github.com", filepath.FromSlash(name), ".git"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	var runErr error
	_, _, err := captureWithInput([]string{"y", "y"}, func() {
		runErr = newApp().Run(context.Background(), append([]string{"ghq", "rm"}, names...))
	})
	if err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatal(runErr)
	}
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, "github.com", filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Errorf("repository %s was not removed: %v", name, err)
		}
	}
}

func TestRmLocalGitSuffix(t *testing.T) {
	for _, tc := range []struct{ ref, bareFlag, want, only string }{
		{ref: "github.com/alice/proj.git", want: "proj.git"},
		{ref: "github.com/alice/proj.git/", want: "proj.git"},
		{ref: "alice/proj.git", want: "proj.git"},
		{ref: "github.com/alice/proj.git", want: "proj.git", only: "proj.git"},
		{ref: "github.com/alice/proj", want: "proj"},
		{ref: "github.com/alice/proj", bareFlag: "true", want: "proj.git"},
		{ref: "github.com/alice/proj.git", bareFlag: "true", want: "proj.git"},
		{ref: "https://github.com/alice/proj.git", bareFlag: "false", want: "proj"},
		{ref: "ssh://git@github.com/alice/proj.git", bareFlag: "false", want: "proj"},
		{ref: "git@github.com:alice/proj.git", bareFlag: "false", want: "proj"},
		{ref: "https://github.com/alice/proj.git", bareFlag: "true", want: "proj.git"},
		{ref: "ssh://git@github.com/alice/proj.git", bareFlag: "true", want: "proj.git"},
		{ref: "git@github.com:alice/proj.git", bareFlag: "true", want: "proj.git"},
		{ref: "https://github.com/alice/proj.git", want: "proj", only: "proj"},
		{ref: "https://github.com/alice/proj.git", bareFlag: "true", want: "proj.git", only: "proj.git"},
	} {
		t.Run(tc.ref+"/"+tc.bareFlag+"/"+tc.only, func(t *testing.T) {
			root := newTempDir(t)
			setRepositoryPathTestRoot(t, root)
			parent := filepath.Join(root, "github.com", "alice")
			names := []string{"proj", "proj.git"}
			if tc.only != "" {
				names = []string{tc.only}
			}
			for _, name := range names {
				initRmRepository(t, filepath.Join(parent, name), name == "proj.git")
			}
			args := []string{"--dry-run"}
			if tc.bareFlag != "" {
				args = append(args, "--bare="+tc.bareFlag)
			}
			target := filepath.Join(parent, tc.want)
			out, stderr, err := runRmTest(t, "y", append(args, tc.ref)...)
			if err != nil || out != "Would remove "+target+"\n" || strings.Contains(stderr, "[y/N]") {
				t.Errorf("dry-run: output %q, stderr %q, error %v; want target %q without confirmation", out, stderr, err, target)
			}
			for _, name := range names {
				assertRmRepositoryKept(t, filepath.Join(parent, name))
			}
		})
	}
}

func TestRmMissingLayout(t *testing.T) {
	for _, tc := range []struct{ ref, existing, missing string }{
		{ref: "example.com/alice/proj.git", existing: "proj", missing: "proj.git"},
		{ref: "https://example.com/alice/proj.git", existing: "proj.git", missing: "proj"},
	} {
		t.Run(tc.ref, func(t *testing.T) {
			root := newTempDir(t)
			setRepositoryPathTestRoot(t, root)
			parent := filepath.Join(root, "example.com", "alice")
			existing := filepath.Join(parent, tc.existing)
			initRmRepository(t, existing, tc.existing == "proj.git")
			wantErr := fmt.Sprintf("directory %q does not exist", filepath.Join(parent, tc.missing))
			for _, args := range [][]string{{"--dry-run", tc.ref}, {tc.ref}} {
				out, stderr, err := runRmTest(t, "y", args...)
				if err == nil || err.Error() != wantErr || out != "" || strings.Contains(stderr, "[y/N]") {
					t.Errorf("args %v: output %q, stderr %q, error %v; want %q before confirmation", args, out, stderr, err, wantErr)
				}
				assertRmRepositoryKept(t, existing)
			}
		})
	}
}

func TestRmLocalGitSuffixRemoval(t *testing.T) {
	root := newTempDir(t)
	setRepositoryPathTestRoot(t, root)
	normalPath := filepath.Join(root, "example.com", "alice", "proj")
	barePath := normalPath + ".git"
	initRmRepository(t, normalPath, false)
	initRmRepository(t, barePath, true)
	for _, response := range []string{"n", "y"} {
		out, stderr, err := runRmTest(t, response, "example.com/alice/proj.git")
		if !strings.Contains(stderr, "Remove "+barePath+"? [y/N]") {
			t.Errorf("confirmation = %q; want target %q", stderr, barePath)
		}
		if response == "n" {
			if err == nil || err.Error() != "aborted" || out != "" {
				t.Errorf("cancellation: output %q, error %v", out, err)
			}
			assertRmRepositoryKept(t, barePath)
		} else {
			if err != nil || out != "Removed "+barePath+"\n" {
				t.Errorf("removal: output %q, error %v; want target %q", out, err, barePath)
			}
			if _, err := os.Stat(barePath); !os.IsNotExist(err) {
				t.Errorf("target was not removed: %v", err)
			}
		}
		assertRmRepositoryKept(t, normalPath)
	}
}

func TestRmRelativePath(t *testing.T) {
	for _, layout := range []string{"normal", "bare"} {
		for _, ref := range []string{".", "./"} {
			t.Run(layout+"/"+ref, func(t *testing.T) {
				root := newTempDir(t)
				setRepositoryPathTestRoot(t, root)
				normalPath := filepath.Join(root, "example.com", "alice", "proj")
				barePath := normalPath + ".git"
				initRmRepository(t, normalPath, false)
				initRmRepository(t, barePath, true)
				target := normalPath
				if layout == "bare" {
					target = barePath
				}
				t.Chdir(target)
				out, stderr, err := runRmTest(t, "y", "--dry-run", ref)
				if err != nil || out != "Would remove "+target+"\n" || strings.Contains(stderr, "[y/N]") {
					t.Errorf("dry-run: output %q, stderr %q, error %v; want target %q without confirmation", out, stderr, err, target)
				}
				for _, path := range []string{normalPath, barePath} {
					assertRmRepositoryKept(t, path)
					if out, err := exec.Command("git", "-C", path, "rev-parse", "--git-dir").CombinedOutput(); err != nil {
						t.Errorf("repository %q is no longer usable: %v\n%s", path, err, out)
					}
				}
			})
		}
	}
}

func runRmTest(t *testing.T, response string, args ...string) (string, string, error) {
	t.Helper()
	var runErr error
	out, stderr, err := captureWithInput([]string{response}, func() {
		// IsSet persists on reused flags, so each run needs a fresh command.
		app := newApp()
		app.Commands = []*cli.Command{newRmCommand()}
		runErr = app.Run(context.Background(), append([]string{"ghq", "rm"}, args...))
	})
	if err != nil {
		t.Fatal(err)
	}
	return out, stderr, runErr
}

func initRmRepository(t *testing.T, path string, bare bool) {
	t.Helper()
	args := []string{"init", "--quiet", "--template="}
	if bare {
		args = append(args, "--bare")
	}
	if out, err := exec.Command("git", append(args, path)...).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(path, "keep"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertRmRepositoryKept(t *testing.T, path string) {
	t.Helper()
	if data, err := os.ReadFile(filepath.Join(path, "keep")); err != nil || string(data) != "keep" {
		t.Errorf("repository %q was changed: %q, %v", path, data, err)
	}
}

func TestRmAmbiguousURL(t *testing.T) {
	for _, tc := range []struct{ ref, layout string }{
		{ref: "https://example.com/alice/proj.git", layout: "same root"},
		{ref: "https://example.com/alice/proj", layout: "same root"},
		{ref: "https://example.com/alice/proj.git/", layout: "same root"},
		{ref: "ssh://git@example.com/alice/proj.git", layout: "same root"},
		{ref: "ssh://git@example.com/alice/proj", layout: "same root"},
		{ref: "git@example.com:alice/proj.git", layout: "same root"},
		{ref: "git@example.com:alice/proj", layout: "same root"},
		{ref: "https://example.com/alice/proj.git", layout: "different roots"},
		{ref: "https://example.com/alice/proj.git", layout: "empty last root"},
		{ref: "https://example.com/alice/proj.git", layout: "empty first root"},
	} {
		t.Run(tc.ref+"/"+tc.layout, func(t *testing.T) {
			base := newTempDir(t)
			normalRoot, bareRoot := filepath.Join(base, "normal"), filepath.Join(base, "normal")
			roots := []string{normalRoot}
			if tc.layout == "different roots" {
				bareRoot = filepath.Join(base, "bare")
				roots = append(roots, bareRoot)
			}
			var emptyPath string
			if tc.layout == "empty last root" || tc.layout == "empty first root" {
				emptyRoot := filepath.Join(base, "empty")
				emptyPath = filepath.Join(emptyRoot, "example.com", "alice", "proj.git")
				if err := os.MkdirAll(emptyPath, 0755); err != nil {
					t.Fatal(err)
				}
				if tc.layout == "empty last root" {
					roots = append(roots, emptyRoot)
				} else {
					roots = append([]string{emptyRoot}, roots...)
				}
			}
			setRepositoryPathTestRoot(t, strings.Join(roots, string(os.PathListSeparator)))
			normalPath := filepath.Join(normalRoot, "example.com", "alice", "proj")
			barePath := filepath.Join(bareRoot, "example.com", "alice", "proj.git")
			initRmRepository(t, normalPath, false)
			initRmRepository(t, barePath, true)
			for _, args := range [][]string{{"--dry-run", tc.ref}, {tc.ref}} {
				out, stderr, err := runRmTest(t, "y", args...)
				if err == nil || !strings.Contains(err.Error(), "ambiguous") {
					t.Errorf("args %v: expected ambiguity error, got output %q, error %v", args, out, err)
				} else {
					for _, fragment := range []string{fmt.Sprintf("%q", normalPath), fmt.Sprintf("%q", barePath), "--bare", "--bare=false"} {
						if !strings.Contains(err.Error(), fragment) {
							t.Errorf("ambiguity error %q does not include %q", err, fragment)
						}
					}
				}
				if out != "" || strings.Contains(stderr, "[y/N]") {
					t.Errorf("ambiguous URL reached output or confirmation: %q, %q", out, stderr)
				}
				assertRmRepositoryKept(t, normalPath)
				assertRmRepositoryKept(t, barePath)
				if emptyPath != "" {
					if fi, err := os.Stat(emptyPath); err != nil || !fi.IsDir() {
						t.Errorf("empty directory was changed: %v", err)
					}
				}
			}
		})
	}
}

func TestRmConflictingBareFlag(t *testing.T) {
	for _, tc := range []struct{ ref, only string }{
		{ref: "github.com/alice/proj.git"},
		{ref: "github.com/alice/proj.git/"},
		{ref: "alice/proj.git"},
		{ref: "github.com/alice/proj.git", only: "proj"},
		{ref: "github.com/alice/proj.git", only: "proj.git"},
	} {
		t.Run(tc.ref+"/"+tc.only, func(t *testing.T) {
			root := newTempDir(t)
			setRepositoryPathTestRoot(t, root)
			parent := filepath.Join(root, "github.com", "alice")
			names := []string{"proj", "proj.git"}
			if tc.only != "" {
				names = []string{tc.only}
			}
			for _, name := range names {
				initRmRepository(t, filepath.Join(parent, name), name == "proj.git")
			}
			for _, args := range [][]string{{"--bare=false", "--dry-run", tc.ref}, {"--bare=false", tc.ref}} {
				out, stderr, err := runRmTest(t, "y", args...)
				if err == nil || !strings.Contains(err.Error(), "conflict") || !strings.Contains(err.Error(), "--bare=false") {
					t.Errorf("args %v: expected conflicting-option error, got %v", args, err)
				}
				if out != "" || strings.Contains(stderr, "[y/N]") {
					t.Errorf("conflicting request reached output or confirmation: %q, %q", out, stderr)
				}
				for _, name := range names {
					assertRmRepositoryKept(t, filepath.Join(parent, name))
				}
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
				if err := os.MkdirAll(filepath.Join(tmpd, "github.com", "motemen", "ghqq", ".git"), 0755); err != nil {
					t.Fatal(err)
				}
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
			input: []string{"rm", "--dry-run", "motemen/ghq-notpermitted"},
			setup: func(t *testing.T) {
				f := filepath.Join(tmpd, "github.com", "motemen", "ghq-notpermitted")
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

			var runErr error
			_, _, err := capture(func() {
				a := newApp()
				args := append([]string{"ghq"}, tc.input...)
				runErr = a.Run(context.Background(), args)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotErr := runErr != nil; gotErr != tc.expectErr {
				t.Fatalf("error = %v, expectErr = %v", runErr, tc.expectErr)
			}
		})
	}
}

func TestRmWorktree(t *testing.T) {
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error { return nil }

	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}
	tmpd := newTempDir(t)
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	setEnv(t, envGhqRoot, tmpd)
	_localRepositoryRoots = nil
	localRepoOnce = &sync.Once{}

	t.Run("rm_linked_worktree", func(t *testing.T) {
		// Create main repo inside ghq root
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-rm", "main"),
			"https://github.com/wt-rm/main.git")

		// Create worktree registered under ghq root so ghq rm can resolve it
		wtDir := filepath.Join(tmpd, "github.com", "wt-rm", "wt-linked")
		addWorktree(t, mainDir, wtDir, "wt-rm-branch")

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-rm/wt-linked"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Worktree directory should be gone
		if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
			t.Error("worktree directory should be removed")
		}

		// Parent repo's .git/worktrees/<name> should be cleaned up
		wtEntry := filepath.Join(mainDir, ".git", "worktrees", "wt-linked")
		if _, err := os.Stat(wtEntry); !os.IsNotExist(err) {
			t.Error("parent repo's worktree entry should be cleaned up")
		}

		// Parent repo should still work
		c := exec.Command("git", "status")
		c.Dir = mainDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Errorf("git status in parent repo failed: %v\n%s", err, out)
		}
	})

	t.Run("rm_dryrun_worktree", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-dry", "main"),
			"https://github.com/wt-dry/main.git")
		wtDir := filepath.Join(tmpd, "github.com", "wt-dry", "wt-linked")
		addWorktree(t, mainDir, wtDir, "wt-dry-branch")

		out, _, err := capture(func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "--dry-run", "wt-dry/wt-linked"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Would remove worktree") {
			t.Errorf("expected 'Would remove worktree' in output, got: %s", out)
		}
		if _, err := os.Stat(wtDir); os.IsNotExist(err) {
			t.Error("worktree should still exist after dry-run")
		}
	})

	t.Run("rm_repo_with_linked_worktrees", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-parent", "repo"),
			"https://github.com/wt-parent/repo.git")

		// Create two external worktrees
		wt1 := filepath.Join(tmpd, "external-wt1")
		wt2 := filepath.Join(tmpd, "external-wt2")
		addWorktree(t, mainDir, wt1, "branch1")
		addWorktree(t, mainDir, wt2, "branch2")

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-parent/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Main repo should be gone
		if _, err := os.Stat(mainDir); !os.IsNotExist(err) {
			t.Error("main repo should be removed")
		}

		// Both worktree directories should be gone
		if _, err := os.Stat(wt1); !os.IsNotExist(err) {
			t.Error("worktree 1 should be removed")
		}
		if _, err := os.Stat(wt2); !os.IsNotExist(err) {
			t.Error("worktree 2 should be removed")
		}
	})

	t.Run("rm_dryrun_with_linked_worktrees", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-dry2", "repo"),
			"https://github.com/wt-dry2/repo.git")

		wt1 := filepath.Join(tmpd, "dry-wt1")
		addWorktree(t, mainDir, wt1, "dry-branch1")

		out, _, err := capture(func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "--dry-run", "wt-dry2/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "linked worktree") {
			t.Errorf("expected 'linked worktree' in output, got: %s", out)
		}
		if _, err := os.Stat(mainDir); os.IsNotExist(err) {
			t.Error("repo should still exist after dry-run")
		}
	})

	t.Run("rm_repo_with_already_deleted_worktree", func(t *testing.T) {
		mainDir := initGitRepo(t, filepath.Join(tmpd, "github.com", "wt-gone", "repo"),
			"https://github.com/wt-gone/repo.git")

		wt := filepath.Join(tmpd, "gone-wt")
		addWorktree(t, mainDir, wt, "gone-branch")

		// Manually delete the worktree directory (simulating user deleting it)
		os.RemoveAll(wt)

		_, _, err := captureWithInput([]string{"y"}, func() {
			a := newApp()
			if e := a.Run(context.Background(), []string{"ghq", "rm", "wt-gone/repo"}); e != nil {
				t.Fatal(e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		// Main repo should be gone regardless
		if _, err := os.Stat(mainDir); !os.IsNotExist(err) {
			t.Error("main repo should be removed even with pre-deleted worktree")
		}
	})
}

func TestRmRepositoryPathCollapse(t *testing.T) {
	type testCase struct {
		ref, path string
		bare      bool
	}
	cases := []testCase{
		{ref: "https://github.com", path: "github.com"},
		{ref: "https://example.com", path: "example.com"},
		{ref: "https://github.com/owner", path: "github.com/owner"},
		{ref: "https://github.com:443/owner", path: "github.com/owner"},
		{ref: "https://GITHUB.COM/owner", path: "GITHUB.COM/owner"},
		{ref: "https://github.com./owner", path: "github.com./owner"},
		{ref: "https://github.com/.git", path: "github.com"},
		{ref: "https://github.com/owner/.git", path: "github.com/owner.git", bare: true},
		{ref: "https://github.com/../../outside", path: "../outside"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases, testCase{ref: "https://github.com/%2e%2e%20/%2e%2e%20/outside", path: "../outside"})
	}
	for _, tc := range cases {
		t.Run(tc.ref, func(t *testing.T) {
			root := filepath.Join(newTempDir(t), "root")
			setRepositoryPathTestRoot(t, root)
			marker := filepath.Join(root, filepath.FromSlash(tc.path), "repo", "keep")
			if err := os.MkdirAll(filepath.Dir(marker), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			args := []string{"ghq", "rm"}
			if tc.bare {
				args = append(args, "--bare")
			}
			var runErr error
			_, stderr, err := captureWithInput([]string{"y"}, func() {
				runErr = newApp().Run(context.Background(), append(args, tc.ref))
			})
			if err != nil {
				t.Fatal(err)
			}
			if runErr == nil || !strings.Contains(runErr.Error(), "repository") {
				t.Errorf("expected repository rejection, got %v", runErr)
			}
			if strings.Contains(stderr, "[y/N]") {
				t.Error("requested confirmation for a collapsed path")
			}
			if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
				t.Errorf("existing repository was changed: %q, %v", data, err)
			}
		})
	}
}
