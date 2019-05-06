package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func withGitConfig(t *testing.T, configContent string) func() {
	tmpdir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}

	tmpGitconfigFile := filepath.Join(tmpdir, "gitconfig")

	ioutil.WriteFile(
		tmpGitconfigFile,
		[]byte(configContent),
		0644,
	)

	prevGitConfigEnv := os.Getenv("GIT_CONFIG")
	os.Setenv("GIT_CONFIG", tmpGitconfigFile)

	return func() {
		os.Setenv("GIT_CONFIG", prevGitConfigEnv)
		os.RemoveAll(tmpdir)
	}
}

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

	rErr, wErr, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	defer wOut.Close()
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

	bufOut, err := ioutil.ReadAll(rOut)
	if err != nil {
		return "", "", err
	}

	bufErr, err := ioutil.ReadAll(rErr)
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
	tmpdir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
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

	return tmpdir
}
