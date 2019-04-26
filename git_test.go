package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestGitConfigAll(t *testing.T) {
	RegisterTestingT(t)

	Expect(GitConfigAll("ghq.non.existent.key")).To(HaveLen(0))
}

func TestGitConfigURL(t *testing.T) {
	RegisterTestingT(t)

	if GitHasFeatureConfigURLMatch() == false {
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

func TestGitVersionOutputSatisfies(t *testing.T) {
	RegisterTestingT(t)

	Expect(
		gitVersionOutputSatisfies(
			"git version 1.7.9",
			[]uint{1, 8, 5},
		),
	).To(BeFalse())

	Expect(
		gitVersionOutputSatisfies(
			"git version 1.8.2.3",
			[]uint{1, 8, 5},
		),
	).To(BeFalse())

	Expect(
		gitVersionOutputSatisfies(
			"git version 1.8.5",
			[]uint{1, 8, 5},
		),
	).To(BeTrue())

	Expect(
		gitVersionOutputSatisfies(
			"git version 1.9.1",
			[]uint{1, 8, 5},
		),
	).To(BeTrue())

	Expect(
		gitVersionOutputSatisfies(
			"git version 2.0.0",
			[]uint{1, 8, 5},
		),
	).To(BeTrue())
}
