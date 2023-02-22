package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Songmu/gitconfig"
)

func samePaths(lhs, rhs string) bool {
	if runtime.GOOS != "windows" {
		return lhs == rhs
	}
	lhss := strings.Split(lhs, "\n")
	rhss := strings.Split(rhs, "\n")
	return samePathSlice(lhss, rhss)
}

func TestDoRoot(t *testing.T) {
	testCases := []struct {
		name              string
		setup             func(t *testing.T)
		expect, allExpect string
		skipOnWin         bool
	}{{
		name: "env",
		setup: func(t *testing.T) {
			setEnv(t, envGhqRoot, "/path/to/ghqroot1"+string(os.PathListSeparator)+"/path/to/ghqroot2")
		},
		expect:    "/path/to/ghqroot1\n",
		allExpect: "/path/to/ghqroot1\n/path/to/ghqroot2\n",
	}, {
		name: "gitconfig",
		setup: func(t *testing.T) {
			setEnv(t, envGhqRoot, "")
			t.Cleanup(gitconfig.WithConfig(t, `
[ghq]
  root = /path/to/ghqroot12
  root = /path/to/ghqroot12
  root = /path/to/ghqroot11
`))
		},
		expect:    "/path/to/ghqroot11\n",
		allExpect: "/path/to/ghqroot11\n/path/to/ghqroot12\n",
		/*
			If your gitconfig contains a path to the start of slash, and you get it with `git config --type=path`,
			the behavior on Windows is strange. Specifically, on Windows with GitHub Actions, a Git
			installation path such as "C:/Program Files/Git/mingw64" is appended immediately before the path.
			This has been addressed in the following issue, which seems to have been resolved in the v2.34.0
			release.
			    https://github.com/git-for-windows/git/pull/3472
			However, Git on GitHub Actions is v2.39.2 at the time of this comment, and this problem continues
			to occur. I'm not sure, so I'll skip the test for now.
		*/
		skipOnWin: true,
	}, {
		name: "default home",
		setup: func(t *testing.T) {
			tmpd := newTempDir(t)
			fpath := filepath.Join(tmpd, "unknown-ghq-dummy")
			f, err := os.Create(fpath)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()

			setEnv(t, envGhqRoot, "")
			setEnv(t, "GIT_CONFIG", fpath)
			setEnv(t, "HOME", "/path/to/ghqhome")
			setEnv(t, "USERPROFILE", "/path/to/ghqhome")
		},
		expect:    "/path/to/ghqhome/ghq\n",
		allExpect: "/path/to/ghqhome/ghq\n",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipOnWin && runtime.GOOS == "windows" {
				t.SkipNow()
			}
			defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
			_localRepositoryRoots = nil
			localRepoOnce = &sync.Once{}
			defer func(orig string) { _home = orig }(_home)
			_home = ""
			homeOnce = &sync.Once{}
			tc.setup(t)
			out, _, _ := capture(func() {
				newApp().Run([]string{"", "root"})
			})
			if !samePaths(out, tc.expect) {
				t.Errorf("got: %s, expect: %s", out, tc.expect)
			}
			out, _, _ = capture(func() {
				newApp().Run([]string{"", "root", "--all"})
			})
			if !samePaths(out, tc.allExpect) {
				t.Errorf("got: %s, expect: %s", out, tc.allExpect)
			}
		})
	}
}
