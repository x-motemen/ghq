package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Songmu/gitconfig"
)

func TestNewURL(t *testing.T) {
	testCases := []struct {
		name, url, expect, host string
		setup                   func() func()
	}{{
		name:   "https", // Does nothing when the URL has scheme part
		url:    "https://github.com/motemen/pusheen-explorer",
		expect: "https://github.com/motemen/pusheen-explorer",
		host:   "github.com",
	}, {
		name:   "scp", // Convert SCP-like URL to SSH URL
		url:    "git@github.com:motemen/pusheen-explorer.git",
		expect: "ssh://git@github.com/motemen/pusheen-explorer.git",
		host:   "github.com",
	}, {
		name:   "scp with root",
		url:    "git@github.com:/motemen/pusheen-explorer.git",
		expect: "ssh://git@github.com/motemen/pusheen-explorer.git",
		host:   "github.com",
	}, {
		name:   "scp without user",
		url:    "github.com:motemen/pusheen-explorer.git",
		expect: "ssh://github.com/motemen/pusheen-explorer.git",
		host:   "github.com",
	}, {
		name:   "different name repository",
		url:    "motemen/ghq",
		expect: "https://github.com/motemen/ghq",
		host:   "github.com",
	}, {
		name:   "with authority repository",
		url:    "github.com/motemen/gore",
		expect: "https://github.com/motemen/gore",
		host:   "github.com",
	}, {
		name:   "with authority repository and go-import",
		url:    "golang.org/x/crypto",
		expect: "https://golang.org/x/crypto",
		host:   "golang.org",
	}, {
		name: "fill username",
		setup: func() func() {
			key := "GITHUB_USER"
			orig := os.Getenv(key)
			os.Setenv(key, "ghq-test")
			return func() { os.Setenv(key, orig) }
		},
		url:    "same-name-ghq",
		expect: "https://github.com/ghq-test/same-name-ghq",
		host:   "github.com",
	}, {
		name: "same name repository",
		setup: func() func() {
			return gitconfig.WithConfig(t, `[ghq]
completeUser = false`)
		},
		url:    "peco",
		expect: "https://github.com/peco/peco",
		host:   "github.com",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				defer tc.setup()()
			}
			repo, err := newURL(tc.url)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if repo.String() != tc.expect {
				t.Errorf("url: got: %s, expect: %s", repo.String(), tc.expect)
			}
			if repo.Host != tc.host {
				t.Errorf("host: got: %s, expect: %s", repo.Host, tc.host)
			}
		})
	}
}

func TestConvertGitURLHTTPToSSH(t *testing.T) {
	testCases := []struct {
		url, expect string
	}{{
		url:    "https://github.com/motemen/pusheen-explorer",
		expect: "ssh://git@github.com/motemen/pusheen-explorer",
	}, {
		url:    "https://ghe.example.com/motemen/pusheen-explorer",
		expect: "ssh://git@ghe.example.com/motemen/pusheen-explorer",
	}, {
		url:    "https://motemen@ghe.example.com/motemen/pusheen-explorer",
		expect: "ssh://motemen@ghe.example.com/motemen/pusheen-explorer",
	}}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			httpsURL, err := newURL(tc.url)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			sshURL, err := convertGitURLHTTPToSSH(httpsURL)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			if sshURL.String() != tc.expect {
				t.Errorf("got: %s, expect: %s", sshURL.String(), tc.expect)
			}
		})
	}
}

func TestNewURL_err(t *testing.T) {
	invalidURL := "http://foo.com/?foo\nbar"
	_, err := newURL(invalidURL)
	const wantSub = "net/url: invalid control character in URL"
	if got := fmt.Sprint(err); !strings.Contains(got, wantSub) {
		t.Errorf("newURL(%q) error = %q; want substring %q", invalidURL, got, wantSub)
	}
	defer gitconfig.WithConfig(t, `[[[`)()

	var exitError *exec.ExitError
	_, err = newURL("peco")
	if !errors.As(err, &exitError) {
		t.Errorf("error should be occurred but nil")
	}
}

func TestFillUsernameToPath_err(t *testing.T) {
	for _, envStr := range []string{"GITHUB_USER", "GITHUB_TOKEN", "USER", "USERNAME"} {
		defer tmpEnv(envStr, "")()
	}
	defer tmpEnv("XDG_CONFIG_HOME", "/dummy/dummy")()
	defer gitconfig.WithConfig(t, "")()

	usr, err := fillUsernameToPath("peco", false)
	t.Log(usr)
	const wantSub = "set ghq.user to your gitconfig"
	if got := fmt.Sprint(err); !strings.Contains(got, wantSub) {
		t.Errorf("fillUsernameToPath(peco, false) error = %q; want substring %q", got, wantSub)
	}
}
