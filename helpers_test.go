package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"testing"
)

func mustParseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return u
}

func captureReader(block func()) (*os.File, *os.File, error) {
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	defer wOut.Close()

	rErr, wErr, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	defer wErr.Close()

	var stdout, stderr *os.File
	os.Stdout, stdout = wOut, os.Stdout
	os.Stderr, stderr = wErr, os.Stderr

	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	block()

	wOut.Close()
	wErr.Close()

	return rOut, rErr, nil
}

func capture(block func()) (string, string, error) {
	rOut, rErr, err := captureReader(block)
	if err != nil {
		return "", "", err
	}
	defer rOut.Close()
	defer rErr.Close()

	bufOut, err := io.ReadAll(rOut)
	if err != nil {
		return "", "", err
	}

	bufErr, err := io.ReadAll(rErr)
	if err != nil {
		return "", "", err
	}

	return string(bufOut), string(bufErr), nil
}

func captureWithInput(in []string, block func()) (string, string, error) {
	rIn, wIn, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	defer rIn.Close()

	var stdin *os.File
	os.Stdin, stdin = rIn, os.Stdin

	defer func() { os.Stdin = stdin }()

	for _, line := range in {
		fmt.Fprintln(wIn, line)
	}
	wIn.Close()
	return capture(block)
}

func newTempDir(t *testing.T) string {
	tmpdir, err := os.MkdirTemp("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpdir) })

	// Resolve /var/folders/.../T/... to /private/var/... in OSX
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %s", err)
	}

	defer func() {
		err := os.Chdir(wd)
		if err != nil {
			t.Fatalf("os.Chdir(): %s", err)
		}
	}()

	err = os.Chdir(tmpdir)
	if err != nil {
		t.Fatalf("os.Chdir(): %s", err)
	}

	tmpdir, err = os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %s", err)
	}

	s, err := toFullPath(tmpdir)
	if err != nil {
		t.Fatalf("toFullPath(%q): %s", tmpdir, err)
	}
	return s
}

func setEnv(t *testing.T, key, val string) {
	orig, ok := os.LookupEnv(key)
	os.Setenv(key, val)

	t.Cleanup(func() {
		if ok {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

// initGitRepo creates a git repo at dir with the given remote URL and an
// initial empty commit. It returns dir for convenience.
func initGitRepo(t *testing.T, dir, remoteURL string) string {
	t.Helper()
	os.MkdirAll(dir, 0755)

	for _, args := range [][]string{
		{"init"},
		{"remote", "add", "origin", remoteURL},
		{"-c", "user.name=test", "-c", "user.email=test@test.com",
			"commit", "--allow-empty", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", args[0], err, out)
		}
	}
	return dir
}

// addWorktree creates a git worktree at wtDir branching from the repo at repoDir.
func addWorktree(t *testing.T, repoDir, wtDir, branch string) {
	t.Helper()
	c := exec.Command("git", "worktree", "add", "-b", branch, wtDir)
	c.Dir = repoDir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}
}
