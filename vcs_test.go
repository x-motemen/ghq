package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/motemen/ghq/cmdutil"
	. "github.com/onsi/gomega"
)

var remoteDummyURL = mustParseURL("https://example.com/git/repo")

func TestGitBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = GitBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "clone", remoteDummyURL.String(), localDir,
	}))

	err = GitBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "clone", "--depth", "1", remoteDummyURL.String(), localDir,
	}))

	err = GitBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "pull", "--ff-only",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}

func TestSubversionBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = SubversionBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"svn", "checkout", remoteDummyURL.String(), localDir,
	}))

	err = SubversionBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"svn", "checkout", "--depth", "1", remoteDummyURL.String(), localDir,
	}))

	err = SubversionBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"svn", "update",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}

func TestGitsvnBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = GitsvnBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "svn", "clone", remoteDummyURL.String(), localDir,
	}))

	err = GitsvnBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "svn", "clone", remoteDummyURL.String(), localDir,
	}))
	err = GitsvnBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"git", "svn", "rebase",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}

func TestMercurialBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = MercurialBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"hg", "clone", remoteDummyURL.String(), localDir,
	}))

	err = MercurialBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"hg", "clone", remoteDummyURL.String(), localDir,
	}))
	err = MercurialBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"hg", "pull", "--update",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}

func TestDarcsBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = DarcsBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"darcs", "get", remoteDummyURL.String(), localDir,
	}))

	err = DarcsBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"darcs", "get", "--lazy", remoteDummyURL.String(), localDir,
	}))

	err = DarcsBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"darcs", "pull",
	}))
}

func TestCvsDummyBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	err = cvsDummyBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).To(HaveOccurred())

	err = cvsDummyBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).To(HaveOccurred())

	err = cvsDummyBackend.Update(localDir, false)

	Expect(err).To(HaveOccurred())
}

func TestBazaarBackend(t *testing.T) {
	RegisterTestingT(t)

	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	err = BazaarBackend.Clone(remoteDummyURL, localDir, false, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(1))
	Expect(lastCommand().Args).To(Equal([]string{
		"bzr", "branch", remoteDummyURL.String(), localDir,
	}))

	err = BazaarBackend.Clone(remoteDummyURL, localDir, true, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(2))
	Expect(lastCommand().Args).To(Equal([]string{
		"bzr", "branch", remoteDummyURL.String(), localDir,
	}))

	err = BazaarBackend.Update(localDir, false)

	Expect(err).NotTo(HaveOccurred())
	Expect(commands).To(HaveLen(3))
	Expect(lastCommand().Args).To(Equal([]string{
		"bzr", "pull",
	}))
	Expect(lastCommand().Dir).To(Equal(localDir))
}
