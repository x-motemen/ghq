package main

import (
	. "github.com/onsi/gomega"
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

	scpUrlWithoutUser, err := NewURL("github.com:motemen/pusheen-explorer.git")
	Expect(scpUrlWithoutUser.String()).To(Equal("ssh://github.com/motemen/pusheen-explorer.git"))
	Expect(scpUrlWithoutUser.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())
}

func TestConvertGitHubURLHTTPToSSH(t *testing.T) {
	RegisterTestingT(t)

	httpsURL, err := NewURL("https://github.com/motemen/pusheen-explorer")
	sshURL, err := ConvertGitHubURLHTTPToSSH(httpsURL)
	Expect(err).To(BeNil())
	Expect(sshURL.String()).To(Equal("ssh://git@github.com/motemen/pusheen-explorer"))
}
