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
	return strings.ToLower(lhs) == strings.ToLower(lhs)
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
		setup             func() func()
		expect, allExpect string
	}{{
		name: "env",
		setup: func() func() {
			orig := os.Getenv(envGhqRoot)
			os.Setenv(envGhqRoot, "/path/to/ghqroot1"+string(os.PathListSeparator)+"/path/to/ghqroot2")
			return func() { os.Setenv(envGhqRoot, orig) }
		},
		expect:    "/path/to/ghqroot1\n",
		allExpect: "/path/to/ghqroot1\n/path/to/ghqroot2\n",
	}, {
		name: "gitconfig",
		setup: func() func() {
			orig := os.Getenv(envGhqRoot)
			os.Setenv(envGhqRoot, "")
			teardown := gitconfig.WithConfig(t, `
[ghq]
  root = /path/to/ghqroot12
  root = /path/to/ghqroot12
  root = /path/to/ghqroot11
`)
			return func() {
				os.Setenv(envGhqRoot, orig)
				teardown()
			}
		},
		expect:    "/path/to/ghqroot11\n",
		allExpect: "/path/to/ghqroot11\n/path/to/ghqroot12\n",
	}, {
		name: "default home",
		setup: func() func() {
			tmpd := newTempDir(t)
			fpath := filepath.Join(tmpd, "unknown-ghq-dummy")
			f, err := os.Create(fpath)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()

			restore1 := tmpEnv(envGhqRoot, "")
			restore2 := tmpEnv("GIT_CONFIG", fpath)
			restore3 := tmpEnv("HOME", "/path/to/ghqhome")

			return func() {
				os.RemoveAll(tmpd)
				restore1()
				restore2()
				restore3()
			}
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
			defer tc.setup()()
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
