package main

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

type goImportTestRepository struct {
	*OtherRepository

	repoURL *url.URL
}

func (repo *goImportTestRepository) VCS() (*VCSBackend, *url.URL, error) {
	return GitBackend, repo.repoURL, nil
}

func TestGetterGoImportCanonicalPathBoundary(t *testing.T) {
	for _, tc := range []struct {
		canonical string
		valid     bool
	}{
		{canonical: "https://example.com", valid: true},
		{canonical: "https://example.com/.git"},
	} {
		t.Run(tc.canonical, func(t *testing.T) {
			withFakeGitBackend(t, func(t *testing.T, root string, clone *_cloneArgs, _ *_updateArgs) {
				t.Setenv(envGhqRoot, root)
				remote := &goImportTestRepository{
					OtherRepository: &OtherRepository{url: mustParseURL("https://example.com/foo")},
					repoURL:         mustParseURL(tc.canonical),
				}
				g := getter{}
				_, err := g.getRemoteRepository(context.Background(), remote, "")
				if !tc.valid {
					if err == nil || clone.local != "" {
						t.Fatalf("unsafe canonical URL: error %v, clone directory %q", err, clone.local)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if want := filepath.Join(root, "example.com", "foo"); clone.local != want {
					t.Errorf("clone directory = %q, want %q", clone.local, want)
				}
			})
		})
	}
}

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
		repoPath:   "/path/to/repo.git",
		expect:     "/path/to/repo",
	}, {
		name:       "trailing slash",
		remotePath: "/path/to/repo/",
		repoPath:   "/path/to/repo/",
		expect:     "/path/to/repo",
	}, {
		name:       ".git/ at the end",
		remotePath: "/path/to/repo.git/",
		repoPath:   "/path/to/repo.git/",
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

func TestGetterCanonicalPathBoundary(t *testing.T) {
	for _, tc := range []struct{ ref, want string }{
		{ref: "https://github.com/owner/.git/tree/main"},
		{ref: "https://github.com/a//b/"},
		{ref: "https://github.com/owner/../repo.git"},
		{ref: "https://github.com/owner/repo/tree/main", want: "github.com/owner/repo"},
	} {
		t.Run(tc.ref, func(t *testing.T) {
			withFakeGitBackend(t, func(t *testing.T, root string, clone *_cloneArgs, _ *_updateArgs) {
				t.Setenv(envGhqRoot, root)
				g := getter{}
				_, err := g.get(context.Background(), tc.ref)
				if tc.want == "" {
					if err == nil || clone.local != "" || !strings.Contains(err.Error(), tc.ref) {
						t.Fatalf("expected rejection identifying input URL: error %v, clone directory %q", err, clone.local)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if want := filepath.Join(root, filepath.FromSlash(tc.want)); clone.local != want {
					t.Errorf("clone directory = %q, want %q", clone.local, want)
				}
			})
		})
	}
}
