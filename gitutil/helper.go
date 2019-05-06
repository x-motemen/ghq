package gitutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// WithConfig is test helper to replace gitconfig temporarily
func WithConfig(t *testing.T, configContent string) func() {
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
