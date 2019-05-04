package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

func WithGitconfigFile(configContent string) (func(), error) {
	tmpdir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		return nil, err
	}

	tmpGitconfigFile := filepath.Join(tmpdir, "gitconfig")

	ioutil.WriteFile(
		tmpGitconfigFile,
		[]byte(configContent),
		0777,
	)

	prevGitConfigEnv := os.Getenv("GIT_CONFIG")
	os.Setenv("GIT_CONFIG", tmpGitconfigFile)

	return func() {
		os.Setenv("GIT_CONFIG", prevGitConfigEnv)
		os.RemoveAll(tmpdir)
	}, nil
}

func mustParseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return u
}
