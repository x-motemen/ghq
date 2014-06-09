package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestGitConfigAll(t *testing.T) {
	RegisterTestingT(t)

	Expect(GitConfigAll("ghq.non.existent.key")).To(HaveLen(0))
}
