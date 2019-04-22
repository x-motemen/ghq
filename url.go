package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// Convert SCP-like URL to SSH URL(e.g. [user@]host.xz:path/to/repo.git/)
// ref. http://git-scm.com/docs/git-fetch#_git_urls
// (golang hasn't supported Perl-like negative look-behind match)
var (
	hasSchemePattern          = regexp.MustCompile("^[^:]+://")
	scpLikeUrlPattern         = regexp.MustCompile("^([^@]+@)?([^:]+):/?(.+)$")
	looksLikeAuthorityPattern = regexp.MustCompile(`[A-Za-z0-9]\.[A-Za-z]+(?::\d{1,5})?`)
)

func NewURL(ref string) (*url.URL, error) {
	if !hasSchemePattern.MatchString(ref) && scpLikeUrlPattern.MatchString(ref) {
		matched := scpLikeUrlPattern.FindStringSubmatch(ref)
		user := matched[1]
		host := matched[2]
		path := matched[3]

		ref = fmt.Sprintf("ssh://%s%s/%s", user, host, path)
	}

	url, err := url.Parse(ref)
	if err != nil {
		return url, err
	}

	if !url.IsAbs() {
		if !strings.Contains(url.Path, "/") {
			url.Path, err = fillUsernameToPath(url.Path)
			if err != nil {
				return url, err
			}
		} else if url.Host == "" {
			// If ref is like "github.com/motemen/ghq" consider it as "https://github.com/motemen/ghq"
			paths := strings.Split(ref, "/")
			if looksLikeAuthorityPattern.MatchString(paths[0]) {
				return url.Parse("https://" + ref)
			}
		}
		url.Scheme = "https"
		url.Host = "github.com"
		if url.Path[0] != '/' {
			url.Path = "/" + url.Path
		}
	}

	return url, nil
}

func ConvertGitURLHTTPToSSH(url *url.URL) (*url.URL, error) {
	sshURL := fmt.Sprintf("ssh://git@%s%s", url.Host, url.Path)
	return url.Parse(sshURL)
}

func fillUsernameToPath(path string) (string, error) {
	user, err := GitConfigSingle("ghq.user")
	if err != nil {
		return path, err
	}
	if user == "" {
		user = os.Getenv("GITHUB_USER")
	}
	if user == "" {
		switch runtime.GOOS {
		case "windows":
			user = os.Getenv("USERNAME")
		default:
			user = os.Getenv("USER")
		}
	}
	if user == "" {
		// Make the error if it does not match any pattern
		return path, fmt.Errorf("set ghq.user to your gitconfig")
	}
	path = user + "/" + path
	return path, nil
}
