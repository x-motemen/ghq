package main

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
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
	Expect(scpUrlWithRoot.String()).To(Equal("ssh://git@github.com//motemen/pusheen-explorer.git"))
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

	withAuthorityRepository, err := NewURL("github.com/motemen/gore")
	Expect(withAuthorityRepository.String()).To(Equal("https://github.com/motemen/gore"))
	Expect(withAuthorityRepository.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	withAuthorityRepository2, err := NewURL("golang.org/x/crypto")
	Expect(withAuthorityRepository2.String()).To(Equal("https://golang.org/x/crypto"))
	Expect(withAuthorityRepository2.Host).To(Equal("golang.org"))
	Expect(err).To(BeNil())

	os.Setenv("GITHUB_USER", "ghq-test")
	sameNameRepository, err := NewURL("same-name-ghq")
	Expect(sameNameRepository.String()).To(Equal("https://github.com/ghq-test/same-name-ghq"))
	Expect(sameNameRepository.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())
}

func TestConvertGitURLHTTPToSSH(t *testing.T) {
	testCases := []struct {
		url, expect string
	}{{
		url:    "https://github.com/motemen/pusheen-explorer",
		expect: "ssh://git@github.com/motemen/pusheen-explorer",
	}, {
		url:    "https://ghe.example.com/motemen/pusheen-explorer",
		expect: "ssh://git@ghe.example.com/motemen/pusheen-explorer",
	}, {
		url:    "https://motemen@ghe.example.com/motemen/pusheen-explorer",
		expect: "ssh://motemen@ghe.example.com/motemen/pusheen-explorer",
	}}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			httpsURL, err := NewURL(tc.url)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			sshURL, err := ConvertGitURLHTTPToSSH(httpsURL)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if sshURL.String() != tc.expect {
				t.Errorf("got: %s, expect: %s", sshURL.String(), tc.expect)
			}
		})
	}
}
