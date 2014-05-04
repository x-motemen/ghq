package main

import (
	"net/url"
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

func TestNewRemoteRepository(t *testing.T) {
	RegisterTestingT(t)

	var (
		repo RemoteRepository
		err  error
	)

	repo, err = NewRemoteRepository(parseURL("https://github.com/motemen/pusheen-explorer"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(true))
	Expect(repo.VCS()).To(Equal(GitBackend))

	repo, err = NewRemoteRepository(parseURL("https://github.com/motemen/pusheen-explorer/blob/master/README.md"))
	Expect(err).To(BeNil())
	Expect(repo.IsValid()).To(Equal(false))

	repo, err = NewRemoteRepository(parseURL("https://example.com/motemen/pusheen-explorer"))
	Expect(err).NotTo(BeNil())
}
