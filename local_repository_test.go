package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLocalRepositoryFromFullPath(t *testing.T) {
	origLocalRepositryRoots := _localRepositoryRoots
	_localRepositoryRoots = []string{"/repos"}
	defer func() { _localRepositoryRoots = origLocalRepositryRoots }()

	testCases := []struct {
		fpath    string
		expect   string
		subpaths []string
	}{{
		fpath:    "/repos/github.com/motemen/ghq",
		expect:   "motemen/ghq",
		subpaths: []string{"ghq", "motemen/ghq", "github.com/motemen/ghq"},
	}, {
		fpath:    "/repos/stash.com/scm/motemen/ghq",
		expect:   "scm/motemen/ghq",
		subpaths: []string{"ghq", "motemen/ghq", "scm/motemen/ghq", "stash.com/scm/motemen/ghq"},
	}}

	for _, tc := range testCases {
		t.Run(tc.fpath, func(t *testing.T) {
			r, err := LocalRepositoryFromFullPath(tc.fpath, nil)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if r.NonHostPath() != tc.expect {
				t.Errorf("NonHostPath: got: %s, expect: %s", r.NonHostPath(), tc.expect)
			}
			if !reflect.DeepEqual(r.Subpaths(), tc.subpaths) {
				t.Errorf("Subpaths:\ngot:    %+v\nexpect: %+v", r.Subpaths(), tc.subpaths)
			}
		})
	}
}

func TestNewLocalRepository(t *testing.T) {
	RegisterTestingT(t)
	_localRepositoryRoots = []string{"/repos"}

	githubURL, _ := url.Parse("ssh://git@github.com/motemen/ghq.git")
	r, err := LocalRepositoryFromURL(githubURL)
	Expect(err).To(BeNil())
	Expect(r.FullPath).To(Equal("/repos/github.com/motemen/ghq"))

	stashURL, _ := url.Parse("ssh://git@stash.com/scm/motemen/ghq")
	r, _ = LocalRepositoryFromURL(stashURL)
	Expect(r.FullPath).To(Equal("/repos/stash.com/scm/motemen/ghq"))

	svnSourceforgeURL, _ := url.Parse("http://svn.code.sf.net/p/ghq/code/trunk")
	r, _ = LocalRepositoryFromURL(svnSourceforgeURL)
	Expect(r.FullPath).To(Equal("/repos/svn.code.sf.net/p/ghq/code/trunk"))

	gitSourceforgeURL, _ := url.Parse("http://git.code.sf.net/p/ghq/code")
	r, _ = LocalRepositoryFromURL(gitSourceforgeURL)
	Expect(r.FullPath).To(Equal("/repos/git.code.sf.net/p/ghq/code"))

	svnSourceforgeJpURL, _ := url.Parse("http://scm.sourceforge.jp/svnroot/ghq/")
	r, _ = LocalRepositoryFromURL(svnSourceforgeJpURL)
	Expect(r.FullPath).To(Equal("/repos/scm.sourceforge.jp/svnroot/ghq"))

	gitSourceforgeJpURL, _ := url.Parse("http://scm.sourceforge.jp/gitroot/ghq/ghq.git")
	r, _ = LocalRepositoryFromURL(gitSourceforgeJpURL)
	Expect(r.FullPath).To(Equal("/repos/scm.sourceforge.jp/gitroot/ghq/ghq"))

	svnAssemblaURL, _ := url.Parse("https://subversion.assembla.com/svn/ghq/")
	r, _ = LocalRepositoryFromURL(svnAssemblaURL)
	Expect(r.FullPath).To(Equal("/repos/subversion.assembla.com/svn/ghq"))

	gitAssemblaURL, _ := url.Parse("https://git.assembla.com/ghq.git")
	r, _ = LocalRepositoryFromURL(gitAssemblaURL)
	Expect(r.FullPath).To(Equal("/repos/git.assembla.com/ghq"))
}

func TestLocalRepositoryRoots(t *testing.T) {
	RegisterTestingT(t)

	defer func(orig string) { os.Setenv("GHQ_ROOT", orig) }(os.Getenv("GHQ_ROOT"))

	_localRepositoryRoots = nil
	os.Setenv("GHQ_ROOT", "/path/to/ghqroot")
	Expect(localRepositoryRoots()).To(Equal([]string{"/path/to/ghqroot"}))

	_localRepositoryRoots = nil
	os.Setenv("GHQ_ROOT", "/path/to/ghqroot1"+string(os.PathListSeparator)+"/path/to/ghqroot2")
	Expect(localRepositoryRoots()).To(Equal([]string{"/path/to/ghqroot1", "/path/to/ghqroot2"}))
}

// https://gist.github.com/kyanny/c231f48e5d08b98ff2c3
func TestList_Symlink(t *testing.T) {
	RegisterTestingT(t)

	root, err := ioutil.TempDir("", "")
	Expect(err).To(BeNil())
	defer os.RemoveAll(root)

	symDir, err := ioutil.TempDir("", "")
	Expect(err).To(BeNil())
	defer os.RemoveAll(symDir)

	_localRepositoryRoots = []string{root}

	err = os.MkdirAll(filepath.Join(root, "github.com", "atom", "atom", ".git"), 0777)
	Expect(err).To(BeNil())

	err = os.MkdirAll(filepath.Join(root, "github.com", "zabbix", "zabbix", ".git"), 0777)
	Expect(err).To(BeNil())

	err = os.Symlink(symDir, filepath.Join(root, "github.com", "ghq"))
	Expect(err).To(BeNil())

	paths := []string{}
	walkLocalRepositories(func(repo *LocalRepository) {
		paths = append(paths, repo.RelPath)
	})

	Expect(paths).To(HaveLen(2))
}
