package main

import . "github.com/onsi/gomega"
import "testing"

func TestParseGitHubURL(t *testing.T) {
	RegisterTestingT(t)

	var (
		u   *GitHubURL
		err error
	)

	u, err = ParseGitHubURL("https://github.com/motemen/pusheen-explorer")
	Expect(err).To(BeNil())
	Expect(u.User).To(Equal("motemen"))
	Expect(u.Repo).To(Equal("pusheen-explorer"))
	Expect(u.Extra).To(Equal(""))

	u, err = ParseGitHubURL("https://github.com/motemen/pusheen-explorer/blob/master/README.md")
	Expect(err).To(BeNil())
	Expect(u.User).To(Equal("motemen"))
	Expect(u.Repo).To(Equal("pusheen-explorer"))
	Expect(u.Extra).To(Equal("blob/master/README.md"))

	u, err = ParseGitHubURL("motemen/pusheen-explorer")
	Expect(err).To(BeNil())
	Expect(u.User).To(Equal("motemen"))
	Expect(u.Repo).To(Equal("pusheen-explorer"))

	u, err = ParseGitHubURL("https://example.com/motemen/pusheen-explorer")
	Expect(err).NotTo(BeNil())
	t.Logf("Got error (successfully): %s", err)
	Expect(u).To(BeNil())

	u, err = ParseGitHubURL("https://github.com/motemen")
	Expect(err).NotTo(BeNil())
	t.Logf("Got error (successfully): %s", err)
	Expect(u).To(BeNil())
}
