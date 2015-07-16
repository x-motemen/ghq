package main

import (
	"errors"
	"net/url"

	"github.com/motemen/ghq/utils"
	. "github.com/onsi/gomega"
)
import "testing"

func parseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return u
}

func TestNewRemoteRepositoryGitHub(t *testing.T) {
	RegisterTestingT(t)

	var (
		repo RemoteRepository
		err  error
	)

	repo, err = NewRemoteRepository(parseURL("https://github.com/motemen/pusheen-explorer"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	Expect(repo.VCS()).To(Equal(GitBackend))

	repo, err = NewRemoteRepository(parseURL("https://github.com/motemen/pusheen-explorer/"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	Expect(repo.VCS()).To(Equal(GitBackend))

	repo, err = NewRemoteRepository(parseURL("https://github.com/motemen/pusheen-explorer/blob/master/README.md"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(false))

	repo, err = NewRemoteRepository(parseURL("https://example.com/motemen/pusheen-explorer"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
}

func TestNewRemoteRepositoryGitHubGist(t *testing.T) {
	RegisterTestingT(t)

	var (
		repo RemoteRepository
		err  error
	)

	repo, err = NewRemoteRepository(parseURL("https://gist.github.com/motemen/9733745"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	Expect(repo.VCS()).To(Equal(GitBackend))
}

func TestNewRemoteRepositoryGoogleCode(t *testing.T) {
	RegisterTestingT(t)

	var (
		repo RemoteRepository
		err  error
	)

	repo, err = NewRemoteRepository(parseURL("https://code.google.com/p/vim/"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	utils.CommandRunner = NewFakeRunner(map[string]error{
		"hg identify":   nil,
		"git ls-remote": errors.New(""),
	})
	Expect(repo.VCS()).To(Equal(MercurialBackend))

	repo, err = NewRemoteRepository(parseURL("https://code.google.com/p/git-core"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	utils.CommandRunner = NewFakeRunner(map[string]error{
		"hg identify":   errors.New(""),
		"git ls-remote": nil,
	})
	Expect(repo.VCS()).To(Equal(GitBackend))
}

func TestNewRemoteRepositoryDarcsHub(t *testing.T) {
	RegisterTestingT(t)

	repo, err := NewRemoteRepository(parseURL("http://hub.darcs.net/foo/bar"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	Expect(repo.VCS()).To(Equal(DarcsBackend))
}
