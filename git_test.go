package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGitConfigAll(t *testing.T) {
	dummyKey := "ghq.non.existent.key"
	confs, err := GitConfigAll(dummyKey)
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	if len(confs) > 0 {
		t.Errorf("GitConfigAll(%q) = %v; want %v", dummyKey, confs, nil)
	}
}

func TestGitConfigURL(t *testing.T) {
	if gitHasFeatureConfigURLMatch() != nil {
		t.Skip("Git does not have config --get-urlmatch feature")
	}

	defer withGitConfig(t, `[ghq "https://ghe.example.com/"]
vcs = github
[ghq "https://ghe.example.com/hg/"]
vcs = hg
`)()

	testCases := []struct {
		name   string
		config []string
		expect string
	}{{
		name:   "github",
		config: []string{"--get-urlmatch", "ghq.vcs", "https://ghe.example.com/foo/bar"},
		expect: "github",
	}, {
		name:   "hg",
		config: []string{"--get-urlmatch", "ghq.vcs", "https://ghe.example.com/hg/repo"},
		expect: "hg",
	}, {
		name:   "empty",
		config: []string{"--get-urlmatch", "ghq.vcs", "https://example.com"},
		expect: "",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := GitConfig(tc.config...)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if value != tc.expect {
				t.Errorf("got: %s, expect: %s", value, tc.expect)
			}
		})
	}
}

func TestGitHasFeatureConfigURLMatch_err(t *testing.T) {
	defer func(orig string) { os.Setenv("PATH", orig) }(os.Getenv("PATH"))
	os.Setenv("PATH", "")

	err := gitHasFeatureConfigURLMatch()
	const wantSub = `failed to execute "git --version": `
	if got := fmt.Sprint(err); !strings.HasPrefix(got, wantSub) {
		t.Errorf("gitHasFeatureConfigURLMatch() error = %q; want substring %q", got, wantSub)
	}
}

func TestGitVersionOutputSatisfies_err(t *testing.T) {
	testCases := []struct {
		in, wantSub string
	}{{
		in:      "brahbrah",
		wantSub: `failed to detect git version from "brahbrah"`,
	}, {
		in:      "18446744073709551616.0.0",
		wantSub: "failed to parse version string",
	}, {
		in:      "1.8.4",
		wantSub: "This version of Git does not support `config --get-urlmatch`; per-URL settings are not available",
	}}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			err := gitVersionOutputSatisfies(tc.in, featureConfigURLMatchVersion)
			if got := fmt.Sprint(err); !strings.HasPrefix(got, tc.wantSub) {
				t.Errorf("gitVersionOutputSatisfies(%s, 1.8.5) error = %q; want substring %q",
					tc.in, got, tc.wantSub)
			}
		})
	}
}
