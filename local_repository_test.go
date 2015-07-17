package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

import (
	"net/url"
	"os"
)

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

	svnSourceforgeURL, _ := url.Parse("http://svn.code.sf.net/p/ghq/code/trunk")
	r = LocalRepositoryFromURL(svnSourceforgeURL)
	Expect(r.FullPath).To(Equal("/repos/svn.code.sf.net/p/ghq/code/trunk"))

	gitSourceforgeURL, _ := url.Parse("http://git.code.sf.net/p/ghq/code")
	r = LocalRepositoryFromURL(gitSourceforgeURL)
	Expect(r.FullPath).To(Equal("/repos/git.code.sf.net/p/ghq/code"))

	svnSourceforgeJpURL, _ := url.Parse("http://scm.sourceforge.jp/svnroot/ghq/")
	r = LocalRepositoryFromURL(svnSourceforgeJpURL)
	Expect(r.FullPath).To(Equal("/repos/scm.sourceforge.jp/svnroot/ghq"))

	gitSourceforgeJpURL, _ := url.Parse("http://scm.sourceforge.jp/gitroot/ghq/ghq.git")
	r = LocalRepositoryFromURL(gitSourceforgeJpURL)
	Expect(r.FullPath).To(Equal("/repos/scm.sourceforge.jp/gitroot/ghq/ghq"))

	svnAssemblaURL, _ := url.Parse("https://subversion.assembla.com/svn/ghq/")
	r = LocalRepositoryFromURL(svnAssemblaURL)
	Expect(r.FullPath).To(Equal("/repos/subversion.assembla.com/svn/ghq"))

	gitAssemblaURL, _ := url.Parse("https://git.assembla.com/ghq.git")
	r = LocalRepositoryFromURL(gitAssemblaURL)
	Expect(r.FullPath).To(Equal("/repos/git.assembla.com/ghq"))
}

func TestLocalRepositoryRoots(t *testing.T) {
	RegisterTestingT(t)

	_localRepositoryRoots = nil

	defer func(orig string) { os.Setenv("GHQ_ROOT", orig) }(os.Getenv("GHQ_ROOT"))
	os.Setenv("GHQ_ROOT", "/path/to/ghqroot")

	Expect(localRepositoryRoots()).To(Equal([]string{"/path/to/ghqroot"}))
}
