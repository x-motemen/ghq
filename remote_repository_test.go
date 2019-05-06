package main

import (
	"testing"
)

func TestNewRemoteRepository(t *testing.T) {
	testCases := []struct {
		url        string
		valid      bool
		vcsBackend *VCSBackend
		repoURL    string
	}{{
		url:        "https://github.com/motemen/pusheen-explorer",
		valid:      true,
		vcsBackend: GitBackend,
	}, {
		url:        "https://github.com/motemen/pusheen-explorer/",
		valid:      true,
		vcsBackend: GitBackend,
	}, {
		url:        "https://github.com/motemen/ghq/logger",
		valid:      true,
		vcsBackend: GitBackend,
		repoURL:    "https://github.com/motemen/ghq",
	}, {
		url:        "https://example.com/motemen/pusheen-explorer/",
		valid:      true,
		vcsBackend: nil,
	}, {
		url:        "https://gist.github.com/motemen/9733745",
		valid:      true,
		vcsBackend: GitBackend,
	}, {
		url:        "http://hub.darcs.net/foo/bar",
		valid:      true,
		vcsBackend: DarcsBackend,
	}}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			repo, err := NewRemoteRepository(mustParseURL(tc.url))
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if repo.IsValid() != tc.valid {
				t.Errorf("repo.IsValid() should be %v, but %v", tc.valid, repo.IsValid())
			}
			vcs, u := repo.VCS()
			if vcs != tc.vcsBackend {
				t.Errorf("got: %+v, expect: %+v", vcs, tc.vcsBackend)
			}
			if tc.repoURL != "" {
				if u.String() != tc.repoURL {
					t.Errorf("repoURL: got: %s, expect: %s", u.String(), tc.repoURL)
				}
			}
		})
	}
}
