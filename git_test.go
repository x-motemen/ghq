package main

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGitConfigAll(t *testing.T) {
	RegisterTestingT(t)

	Expect(GitConfigAll("ghq.non.existent.key")).To(HaveLen(0))
}

func TestGitConfigURL(t *testing.T) {
	RegisterTestingT(t)

	if GitHasFeatureConfigURLMatch() != nil {
		t.Skip("Git does not have config --get-urlmatch feature")
	}

	reset, err := WithGitconfigFile(`
[ghq "https://ghe.example.com/"]
vcs = github
[ghq "https://ghe.example.com/hg/"]
vcs = hg
`)
	if err != nil {
		t.Fatal(err)
	}
	defer reset()

	var (
		value string
	)

	value, err = GitConfig("--get-urlmatch", "ghq.vcs", "https://ghe.example.com/foo/bar")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal("github"))

	value, err = GitConfig("--get-urlmatch", "ghq.vcs", "https://ghe.example.com/hg/repo")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal("hg"))

	value, err = GitConfig("--get-urlmatch", "ghq.vcs", "https://example.com")
	Expect(err).NotTo(HaveOccurred())
	Expect(value).To(Equal(""))
}
