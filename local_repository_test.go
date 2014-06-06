package main

import . "github.com/onsi/gomega"
import "net/url"
import "testing"

func TestNewLocalRepository(t *testing.T) {
	RegisterTestingT(t)

	_localRepositoryRoots = []string{"/repos"}

	r, err := LocalRepositoryFromFullPath("/repos/github.com/motemen/ghq")
	Expect(err).To(BeNil())
	Expect(r.NonHostPath()).To(Equal("motemen/ghq"))
	Expect(r.Subpaths()).To(Equal([]string{"ghq", "motemen/ghq", "github.com/motemen/ghq"}))

	r, err = LocalRepositoryFromFullPath("/repos/stash.com/scm/motemen/ghq")
	Expect(err).To(BeNil())
	Expect(r.NonHostPath()).To(Equal("scm/motemen/ghq"))
	Expect(r.Subpaths()).To(Equal([]string{"ghq", "motemen/ghq", "scm/motemen/ghq", "stash.com/scm/motemen/ghq"}))

	githubURL, _ := url.Parse("ssh://git@github.com/motemen/ghq.git")
	r = LocalRepositoryFromURL(githubURL)
	Expect(r.FullPath).To(Equal("/repos/github.com/motemen/ghq"))

	stashURL, _ := url.Parse("ssh://git@stash.com/scm/motemen/ghq")
	r = LocalRepositoryFromURL(stashURL)
	Expect(r.FullPath).To(Equal("/repos/stash.com/scm/motemen/ghq"))
}
