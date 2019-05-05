package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
	origLocalRepositryRoots := _localRepositoryRoots
	_localRepositoryRoots = []string{"/repos"}
	defer func() { _localRepositoryRoots = origLocalRepositryRoots }()

	testCases := []struct {
		name, url, expect string
	}{{
		name:   "GitHub",
		url:    "ssh://git@github.com/motemen/ghq.git",
		expect: "/repos/github.com/motemen/ghq",
	}, {
		name:   "stash",
		url:    "ssh://git@stash.com/scm/motemen/ghq.git",
		expect: "/repos/stash.com/scm/motemen/ghq",
	}, {
		name:   "svn Sourceforge",
		url:    "http://svn.code.sf.net/p/ghq/code/trunk",
		expect: "/repos/svn.code.sf.net/p/ghq/code/trunk",
	}, {
		name:   "git Sourceforge",
		url:    "http://git.code.sf.net/p/ghq/code",
		expect: "/repos/git.code.sf.net/p/ghq/code",
	}, {
		name:   "svn Sourceforge JP",
		url:    "http://scm.sourceforge.jp/svnroot/ghq/",
		expect: "/repos/scm.sourceforge.jp/svnroot/ghq",
	}, {
		name:   "git Sourceforge JP",
		url:    "http://scm.sourceforge.jp/gitroot/ghq/ghq.git",
		expect: "/repos/scm.sourceforge.jp/gitroot/ghq/ghq",
	}, {
		name:   "svn Assembla",
		url:    "https://subversion.assembla.com/svn/ghq/",
		expect: "/repos/subversion.assembla.com/svn/ghq",
	}, {
		name:   "git Assembla",
		url:    "https://git.assembla.com/ghq.git",
		expect: "/repos/git.assembla.com/ghq",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := LocalRepositoryFromURL(mustParseURL(tc.url))
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if r.FullPath != tc.expect {
				t.Errorf("got: %s, expect: %s", r.FullPath, tc.expect)
			}
		})
	}
}

func TestLocalRepositoryRoots(t *testing.T) {
	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	defer func(orig string) { os.Setenv("GHQ_ROOT", orig) }(os.Getenv("GHQ_ROOT"))

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		root   string
		expect []string
	}{{
		root:   "/path/to/ghqroot",
		expect: []string{"/path/to/ghqroot"},
	}, {
		root:   "/path/to/ghqroot1" + string(os.PathListSeparator) + "/path/to/ghqroot2",
		expect: []string{"/path/to/ghqroot1", "/path/to/ghqroot2"},
	}, {
		root:   "/path/to/ghqroot11" + string(os.PathListSeparator) + "vendor",
		expect: []string{"/path/to/ghqroot11", filepath.Join(wd, "vendor")},
	}}

	for _, tc := range testCases {
		t.Run(tc.root, func(t *testing.T) {
			_localRepositoryRoots = nil
			os.Setenv("GHQ_ROOT", tc.root)
			got, err := localRepositoryRoots()
			if err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}
			if !reflect.DeepEqual(got, tc.expect) {
				t.Errorf("\ngot:    %+v\nexpect: %+v", got, tc.expect)
			}
		})
	}
}

// https://gist.github.com/kyanny/c231f48e5d08b98ff2c3
func TestList_Symlink(t *testing.T) {
	root, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	symDir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(symDir)

	origLocalRepositryRoots := _localRepositoryRoots
	_localRepositoryRoots = []string{root}
	defer func() { _localRepositoryRoots = origLocalRepositryRoots }()

	if err := os.MkdirAll(filepath.Join(root, "github.com", "atom", "atom", ".git"), 0777); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(root, "github.com", "zabbix", "zabbix", ".git"), 0777); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(symDir, filepath.Join(root, "github.com", "ghq")); err != nil {
		t.Fatal(err)
	}

	paths := []string{}
	walkLocalRepositories(func(repo *LocalRepository) {
		paths = append(paths, repo.RelPath)
	})

	if len(paths) != 2 {
		t.Errorf("length of paths should be 2, but: %d", len(paths))
	}
}

func TestFindVCSBackend(t *testing.T) {
	newTmp := func() string {
		tmpdir, err := ioutil.TempDir("", "ghq-test")
		if err != nil {
			t.Fatal(err)
		}
		return tmpdir
	}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) (string, func())
		expect *VCSBackend
	}{{
		name: "git",
		setup: func(t *testing.T) (string, func()) {
			dir := newTmp()
			os.MkdirAll(filepath.Join(dir, ".git"), 0755)
			return dir, func() {
				os.RemoveAll(dir)
			}
		},
		expect: GitBackend,
	}, {
		name: "git svn",
		setup: func(t *testing.T) (string, func()) {
			dir := newTmp()
			os.MkdirAll(filepath.Join(dir, ".git", "svn"), 0755)
			return dir, func() {
				os.RemoveAll(dir)
			}
		},
		expect: GitsvnBackend,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fpath, teardown := tc.setup(t)
			defer teardown()
			backend := findVCSBackend(fpath)
			if backend != tc.expect {
				t.Errorf("got: %v, expect: %v", backend, tc.expect)
			}
		})
	}
}
