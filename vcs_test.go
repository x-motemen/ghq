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

func TestVCSBackend(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	localDir := filepath.Join(tempDir, "repo")
	commands := []*exec.Cmd{}
	lastCommand := func() *exec.Cmd { return commands[len(commands)-1] }
	defer func(orig func(cmd *exec.Cmd) error) {
		cmdutil.CommandRunner = orig
	}(cmdutil.CommandRunner)
	cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
		commands = append(commands, cmd)
		return nil
	}

	testCases := []struct {
		name   string
		f      func() error
		expect []string
		dir    string
	}{{
		name: "[git] clone",
		f: func() error {
			return GitBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"git", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] shallow clone",
		f: func() error {
			return GitBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"git", "clone", "--depth", "1", remoteDummyURL.String(), localDir},
	}, {
		name: "[git] update",
		f: func() error {
			return GitBackend.Update(localDir, false)
		},
		expect: []string{"git", "pull", "--ff-only"},
		dir:    localDir,
	}, {
		name: "[svn] checkout",
		f: func() error {
			return SubversionBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"svn", "checkout", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] checkout shallow",
		f: func() error {
			return SubversionBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"svn", "checkout", "--depth", "1", remoteDummyURL.String(), localDir},
	}, {
		name: "[svn] update",
		f: func() error {
			return SubversionBackend.Update(localDir, false)
		},
		expect: []string{"svn", "update"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone",
		f: func() error {
			return GitsvnBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[git-svn] update",
		f: func() error {
			return GitsvnBackend.Update(localDir, false)
		},
		expect: []string{"git", "svn", "rebase"},
		dir:    localDir,
	}, {
		name: "[git-svn] clone shallow",
		f: func() error {
			return GitsvnBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"git", "svn", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] clone",
		f: func() error {
			return MercurialBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[hg] update",
		f: func() error {
			return MercurialBackend.Update(localDir, false)
		},
		expect: []string{"hg", "pull", "--update"},
		dir:    localDir,
	}, {
		name: "[hg] clone shallow",
		f: func() error {
			return MercurialBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"hg", "clone", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone",
		f: func() error {
			return DarcsBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"darcs", "get", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] clone shallow",
		f: func() error {
			return DarcsBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"darcs", "get", "--lazy", remoteDummyURL.String(), localDir},
	}, {
		name: "[darcs] update",
		f: func() error {
			return DarcsBackend.Update(localDir, false)
		},
		expect: []string{"darcs", "pull"},
		dir:    localDir,
	}, {
		name: "[bzr] clone",
		f: func() error {
			return BazaarBackend.Clone(remoteDummyURL, localDir, false, false)
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
	}, {
		name: "[bzr] update",
		f: func() error {
			return BazaarBackend.Update(localDir, false)
		},
		expect: []string{"bzr", "pull"},
		dir:    localDir,
	}, {
		name: "[bzr] clone shallow",
		f: func() error {
			return BazaarBackend.Clone(remoteDummyURL, localDir, true, false)
		},
		expect: []string{"bzr", "branch", remoteDummyURL.String(), localDir},
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
