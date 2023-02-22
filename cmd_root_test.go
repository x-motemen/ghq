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

func samePath(lhs, rhs string) bool {
	if runtime.GOOS != "windows" {
		return lhs == rhs
	}

	lhs, _ = filepath.Abs(filepath.Clean(lhs))
	rhs, _ = filepath.Abs(filepath.Clean(rhs))
	return strings.ToLower(lhs) == strings.ToLower(rhs)
}

func samePaths(lhs, rhs string) bool {
	if runtime.GOOS != "windows" {
		return lhs == rhs
	}
	lhss := strings.Split(lhs, "\n")
	rhss := strings.Split(rhs, "\n")
	for i := range lhss {
		if !samePath(lhss[i], rhss[i]) {
			return false
		}
	}
	return true
}

func TestDoRoot(t *testing.T) {
	testCases := []struct {
		name              string
		setup             func(t *testing.T)
		expect, allExpect string
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
		},
		expect:    "/path/to/ghqhome/ghq\n",
		allExpect: "/path/to/ghqhome/ghq\n",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
