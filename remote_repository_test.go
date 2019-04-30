package main

import (
	"errors"
	"net/url"

	"testing"

	"github.com/motemen/ghq/cmdutil"
	. "github.com/onsi/gomega"
)

func mustParseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return u
}

func TestNewRemoteRepository(t *testing.T) {
	testCases := []struct {
		url        string
		valid      bool
		vcsBackend *VCSBackend
	}{
		{
			url:        "https://github.com/motemen/pusheen-explorer",
			valid:      true,
			vcsBackend: GitBackend,
		},
		{
			url:        "https://github.com/motemen/pusheen-explorer/",
			valid:      true,
			vcsBackend: GitBackend,
		},
		{
			url:        "https://github.com/motemen/pusheen-explorer/blob/master/README.md",
			valid:      false,
			vcsBackend: GitBackend,
		},
		{
			url:        "https://example.com/motemen/pusheen-explorer/",
			valid:      true,
			vcsBackend: nil,
		},
		{
			url:        "https://gist.github.com/motemen/9733745",
			valid:      true,
			vcsBackend: GitBackend,
		},
		{
			url:        "http://hub.darcs.net/foo/bar",
			valid:      true,
			vcsBackend: DarcsBackend,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			repo, err := NewRemoteRepository(mustParseURL(tc.url))
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if repo.IsValid() != tc.valid {
				t.Errorf("repo.IsValid() should be %v, but %v", tc.valid, repo.IsValid())
			}
			vcs, _ := repo.VCS()
			if vcs != tc.vcsBackend {
				t.Errorf("got: %+v, expect: %+v", vcs, tc.vcsBackend)
			}
		})
	}
}

func TestNewRemoteRepositoryGoogleCode(t *testing.T) {
	RegisterTestingT(t)

	var (
		repo RemoteRepository
		err  error
	)

	repo, err = NewRemoteRepository(mustParseURL("https://code.google.com/p/vim/"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	cmdutil.CommandRunner = NewFakeRunner(map[string]error{
		"hg identify":   nil,
		"git ls-remote": errors.New(""),
	})
	vcs, _ := repo.VCS()
	Expect(vcs).To(Equal(MercurialBackend))

	repo, err = NewRemoteRepository(mustParseURL("https://code.google.com/p/git-core"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	cmdutil.CommandRunner = NewFakeRunner(map[string]error{
		"hg identify":   errors.New(""),
		"git ls-remote": nil,
	})
	vcs, _ = repo.VCS()
	Expect(vcs).To(Equal(GitBackend))
}
