package main

import "testing"

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
	if GitHasFeatureConfigURLMatch() != nil {
		t.Skip("Git does not have config --get-urlmatch feature")
	}

	reset, err := WithGitconfigFile(`
[ghq "https://ghe.example.com/"]
vcs = github
[ghq "https://ghe.example.com/hg/"]
vcs = hg
`)
	if err != nil {
		t.Fatal(err)
	}
	defer reset()

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
