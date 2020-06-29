package main

import "testing"

func TestDetectLocalRepoRoot(t *testing.T) {
	testCases := []struct {
		name, remotePath, repoPath, expect string
	}{{
		name:       "same",
		remotePath: "/motemen/ghq",
		repoPath:   "/motemen/ghq",
		expect:     "/motemen/ghq",
	}, {
		name:       "deep remote repo path",
		remotePath: "/path/to/repo/repo",
		repoPath:   "/path/to/repo",
		expect:     "/path/to/repo",
	}, {
		name:       "different remote root",
		remotePath: "/src/path/to/repo/repo",
		repoPath:   "/path/to/repo",
		expect:     "/src/path/to/repo",
	}, {
		name:       "different repo root",
		remotePath: "/path/to/repo/repo",
		repoPath:   "/git/path/to/repo",
		expect:     "/path/to/repo",
	}, {
		name:       "different roots",
		remotePath: "/src/path/to/repo/repo",
		repoPath:   "/git/path/to/repo",
		expect:     "/src/path/to/repo",
	}, {
		name:       "different roots with multibyte",
		remotePath: "/そーすこーど/path/to/repo/repo",
		repoPath:   "/ぎっと/path/to/repo",
		expect:     "/そーすこーど/path/to/repo",
	}, {
		name:       "shallow path",
		remotePath: "/zap/buffer",
		repoPath:   "/uber-go/zap",
		expect:     "/zap",
	}, {
		name:       ".git at the end",
		remotePath: "/path/to/repo.git",
		repoPath:   "/path/to/repo",
		expect:     "/path/to/repo",
	}, {
		name:       "trailing slash",
		remotePath: "/path/to/repo/",
		repoPath:   "/path/to/repo",
		expect:     "/path/to/repo",
	}, {
		name:       ".git/ at the end",
		remotePath: "/path/to/repo.git/",
		repoPath:   "/path/to/repo",
		expect:     "/path/to/repo",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := detectLocalRepoRoot(tc.remotePath, tc.repoPath)
			if tc.expect != out {
				t.Errorf("detectLocalRepoRoot(%q, %q) = %q, expect: %q",
					tc.remotePath, tc.repoPath, out, tc.expect)
			}
		})
	}
}
