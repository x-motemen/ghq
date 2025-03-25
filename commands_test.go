package main

import (
	"net/url"
	"path/filepath"
	"sync"
	"testing"
)

type _cloneArgs struct {
	remote    *url.URL
	local     string
	shallow   bool
	branch    string
	recursive bool
	bare      bool
	silent    bool
	partial   string
}

type _updateArgs struct {
	local string
}

func withFakeGitBackend(t *testing.T, block func(*testing.T, string, *_cloneArgs, *_updateArgs)) {
	tmpRoot := newTempDir(t)

	defer func(orig []string) { _localRepositoryRoots = orig }(_localRepositoryRoots)
	_localRepositoryRoots = []string{tmpRoot}

	var cloneArgs _cloneArgs
	var updateArgs _updateArgs

	var originalGitBackend = GitBackend
	tmpBackend := &VCSBackend{
		Clone: func(vg *vcsGetOption) error {
			cloneArgs = _cloneArgs{
				remote:    vg.url,
				local:     filepath.FromSlash(vg.dir),
				shallow:   vg.shallow,
				branch:    vg.branch,
				recursive: vg.recursive,
				bare:      vg.bare,
				silent:    vg.silent,
				partial:   vg.partial,
			}
			return nil
		},
		Update: func(vg *vcsGetOption) error {
			updateArgs = _updateArgs{
				local: vg.dir,
			}
			return nil
		},
	}
	defer func(orig string) { _home = orig }(_home)
	_home = ""
	homeOnce = &sync.Once{}

	GitBackend = tmpBackend
	vcsContentsMap[".git"] = tmpBackend
	defer func() { GitBackend = originalGitBackend; vcsContentsMap[".git"] = originalGitBackend }()
	block(t, tmpRoot, &cloneArgs, &updateArgs)
}
