package main

import . "github.com/onsi/gomega"
import "testing"

func TestNewLocalRepository(t *testing.T) {
	RegisterTestingT(t)

	_localRepositoriesRoot = "/repos"

	r, err := LocalRepositoryFromFullPath("/repos/github.com/motemen/ghq")
	Expect(err).To(BeNil())
	Expect(r.NonHostPath()).To(Equal("motemen/ghq"))
	Expect(r.Subpaths()).To(Equal([]string{"github.com/motemen/ghq", "motemen/ghq", "ghq"}))
}
