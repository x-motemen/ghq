package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/motemen/ghq/cmdutil"
	. "github.com/onsi/gomega"
)

var remoteDummyURL = mustParseURL("https://example.com/git/repo")

type vcsBackendTest struct {
	name   string
	f      func() error
	expect []string
	dir    string
}

func vcsTestSetup() (string, func() *exec.Cmd, func()) {
	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		panic(err)
	}

	localDir := filepath.Join(tempDir, "repo")

	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }

	orig := cmdutil.CommandRunner
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	return localDir, lastCommand, func() {
		os.RemoveAll(tempDir)
		cmdutil.CommandRunner = orig
	}
}

func TestGitBackend(t *testing.T) {
	localDir, lastCommand, teardown := vcsTestSetup()
	defer teardown()

	testCases := []vcsBackendTest{{
		name: "clone",
		f: func() error {
			return GitBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"git", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "shallow clone",
		f: func() error {
			return GitBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"git", "clone", "--depth", "1", remoteDummyURL.String(), localDir},
	}, {
		name: "update",
		f: func() error {
			return GitBackend.Update(localDir, false)
		},
		expect: []string{"git", "pull", "--ff-only"},
		dir:    localDir,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.f(); err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}
			c := lastCommand()
			if !reflect.DeepEqual(c.Args, tc.expect) {
				t.Errorf("\ngot:  %+v\nexpect: %+v", c.Args, tc.expect)
			}
			if c.Dir != tc.dir {
				t.Errorf("got: %s, expect: %s", c.Dir, tc.dir)
			}
		})
	}
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
