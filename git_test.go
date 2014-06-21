package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	. "github.com/onsi/gomega"
)

func TestGitConfigAll(t *testing.T) {
	RegisterTestingT(t)

	Expect(GitConfigAll("ghq.non.existent.key")).To(HaveLen(0))
}

func TestGitConfigURL(t *testing.T) {
	RegisterTestingT(t)

	tmpdir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}

	tmpGitconfigFile := filepath.Join(tmpdir, "gitconfig")

	ioutil.WriteFile(
		tmpGitconfigFile,
		[]byte(`
[ghq.url "https://ghe.example.com/"]
vcs = github
[ghq.url "https://ghe.example.com/hg/"]
vcs = hg
		`),
		0777,
	)

	os.Setenv("GIT_CONFIG", tmpGitconfigFile)

	var (
		value string
	)

	value, err = GitConfigURLMatch("ghq.url", "vcs", "https://ghe.example.com/foo/bar")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal("github"))

	value, err = GitConfigURLMatch("ghq.url", "vcs", "https://ghe.example.com/hg/repo")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal("hg"))

	value, err = GitConfigURLMatch("ghq.url", "vcs", "https://example.com")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal(""))
}
