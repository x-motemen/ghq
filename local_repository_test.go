package main

import . "github.com/onsi/gomega"
import "testing"

func TestNewLocalRepository(t *testing.T) {
	RegisterTestingT(t)

	_localRepositoryRoots = []string{"/repos"}

	r, err := LocalRepositoryFromFullPath("/repos/github.com/motemen/ghq")
	Expect(err).To(BeNil())
	Expect(r.NonHostPath()).To(Equal("motemen/ghq"))
	Expect(r.Subpaths()).To(Equal([]string{"ghq", "motemen/ghq", "github.com/motemen/ghq"}))
}
