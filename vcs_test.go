package main

import (
	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/motemen/ghq/utils"
	. "github.com/onsi/gomega"
)

func TestGitBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}

	localDir := filepath.Join(tempDir, "repo")

	remoteURL, err := url.Parse("https://example.com/git/repo")
	if err != nil {
		t.Fatal(err)
	}

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	utils.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = GitBackend.Clone(remoteURL, localDir)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "clone", remoteURL.String(), localDir,
	}))

	err = GitBackend.Update(localDir)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "pull", "--ff-only",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}

func TestMercurialBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}

	localDir := filepath.Join(tempDir, "repo")

	remoteURL, err := url.Parse("https://example.com/git/repo")
	if err != nil {
		t.Fatal(err)
	}

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	utils.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = MercurialBackend.Clone(remoteURL, localDir)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"hg", "clone", remoteURL.String(), localDir,
	}))

	err = MercurialBackend.Update(localDir)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"hg", "pull",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}
