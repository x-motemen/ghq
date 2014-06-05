package main

import (
	"testing"
	. "github.com/onsi/gomega"
)

func TestGitConfigAll(t *testing.T) {
	RegisterTestingT(t)

	Expect(GitConfigAll("ghq.non.existent.key")).To(HaveLen(0))
}
