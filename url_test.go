package main

import (
	. "github.com/onsi/gomega"
	"net/url"
	"testing"
)

func TestNewURL(t *testing.T) {
	RegisterTestingT(t)

	// Does nothing whent the URL has scheme part
	httpsUrl, err := NewURL("https://github.com/motemen/pusheen-explorer")
	Expect(httpsUrl.String()).To(Equal("https://github.com/motemen/pusheen-explorer"))
	Expect(httpsUrl.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	// Convert SCP-like URL to SSH URL
	scpUrl, err := NewURL("git@github.com:motemen/pusheen-explorer.git")
	Expect(scpUrl.String()).To(Equal("ssh://git@github.com/motemen/pusheen-explorer.git"))
	Expect(scpUrl.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	scpUrlWithRoot, err := NewURL("git@github.com:/motemen/pusheen-explorer.git")
	Expect(scpUrlWithRoot.String()).To(Equal("ssh://git@github.com/motemen/pusheen-explorer.git"))
	Expect(scpUrlWithRoot.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	scpUrlWithoutUser, err := NewURL("github.com:motemen/pusheen-explorer.git")
	Expect(scpUrlWithoutUser.String()).To(Equal("ssh://github.com/motemen/pusheen-explorer.git"))
	Expect(scpUrlWithoutUser.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	differentNameRepository, err := NewURL("motemen/ghq")
	Expect(differentNameRepository.String()).To(Equal("https://github.com/motemen/ghq"))
	Expect(differentNameRepository.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	sameNameRepository, err := NewURL("same-name-ghq")
	Expect(sameNameRepository.String()).To(Equal("https://github.com/same-name-ghq/same-name-ghq"))
	Expect(sameNameRepository.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())
}

func TestConvertGitURLHTTPToSSH(t *testing.T) {
	RegisterTestingT(t)

	var (
		httpsURL, sshURL *url.URL
		err              error
	)

	httpsURL, err = NewURL("https://github.com/motemen/pusheen-explorer")
	sshURL, err = ConvertGitURLHTTPToSSH(httpsURL)
	Expect(err).To(BeNil())
	Expect(sshURL.String()).To(Equal("ssh://git@github.com/motemen/pusheen-explorer"))

	httpsURL, err = NewURL("https://ghe.example.com/motemen/pusheen-explorer")
	sshURL, err = ConvertGitURLHTTPToSSH(httpsURL)
	Expect(err).To(BeNil())
	Expect(sshURL.String()).To(Equal("ssh://git@ghe.example.com/motemen/pusheen-explorer"))
}
