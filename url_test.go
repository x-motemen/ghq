package main

import (
	"testing"
	. "github.com/onsi/gomega"
)

func TestNewURL(t *testing.T) {
	RegisterTestingT(t)

	httpsUrl, err := NewURL("https://github.com/motemen/pusheen-explorer")
	Expect(httpsUrl.String()).To(Equal("https://github.com/motemen/pusheen-explorer"))
	Expect(httpsUrl.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())

	sshUrl, err := NewURL("git@github.com:motemen/pusheen-explorer.git")
	Expect(sshUrl.String()).To(Equal("ssh://git@github.com/motemen/pusheen-explorer.git"))
	Expect(sshUrl.Host).To(Equal("github.com"))
	Expect(err).To(BeNil())
}
