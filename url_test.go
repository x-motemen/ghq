package main

import (
	"testing"
	. "github.com/onsi/gomega"
)

func TestNewURL(t *testing.T) {
	url, err := NewURL("https://github.com/motemen/pusheen-explorer")
	Expect(url.String()).To(Equal("https://github.com/motemen/pusheen-explorer"))
	Expect(err).To(BeNil())

	url, err = NewURL("git@github.com:motemen/pusheen-explorer.git")
	Expect(url.Host).To(Equal("ssh://git@github.com/motemen/pusheen-explorer.git"))
	Expect(err).To(BeNil())
}
