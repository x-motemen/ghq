package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/Songmu/gitconfig"
)

// Convert SCP-like URL to SSH URL(e.g. [user@]host.xz:path/to/repo.git/)
// ref. http://git-scm.com/docs/git-fetch#_git_urls
// (golang hasn't supported Perl-like negative look-behind match)
var (
	hasSchemePattern          = regexp.MustCompile("^[^:]+://")
	scpLikeURLPattern         = regexp.MustCompile("^([^@]+@)?([^:]+):(/?.+)$")
	looksLikeAuthorityPattern = regexp.MustCompile(`[A-Za-z0-9]\.[A-Za-z]+(?::\d{1,5})?$`)
)

func newURL(ref string, forceMe bool) (*url.URL, error) {
	if !hasSchemePattern.MatchString(ref) {
		if scpLikeURLPattern.MatchString(ref) {
			matched := scpLikeURLPattern.FindStringSubmatch(ref)
			user := matched[1]
			host := matched[2]
			path := matched[3]
			// If the path is a relative path not beginning with a slash like
			// `path/to/repo`, we might convert to like
			// `ssh://user@repo.example.com/~/path/to/repo` using tilde, but
			// since GitHub doesn't support it, we treat relative and absolute
			// paths the same way.
			ref = fmt.Sprintf("ssh://%s%s/%s", user, host, strings.TrimPrefix(path, "/"))
		} else {
			// If ref is like "github.com/motemen/ghq" convert to "https://github.com/motemen/ghq"
			paths := strings.Split(ref, "/")
			if len(paths) > 1 && looksLikeAuthorityPattern.MatchString(paths[0]) {
				ref = "https://" + ref
			}
		}
	}

	url, err := url.Parse(ref)
	if err != nil {
		return url, err
	}

	if !url.IsAbs() {
		if !strings.Contains(url.Path, "/") {
			url.Path, err = fillUsernameToPath(url.Path, forceMe)
			if err != nil {
				return url, err
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

func convertGitURLHTTPToSSH(url *url.URL) (*url.URL, error) {
	user := "git"
	if url.User != nil {
		user = url.User.Username()
	}
	sshURL := fmt.Sprintf("ssh://%s@%s%s", user, url.Host, url.Path)
	return url.Parse(sshURL)
}

func detectUserName() (string, error) {
	user, err := gitconfig.Get("ghq.user")
	if (err != nil && !gitconfig.IsNotFound(err)) || user != "" {
		return user, err
	}

	user, err = gitconfig.GitHubUser("")
	if (err != nil && !gitconfig.IsNotFound(err)) || user != "" {
		return user, err
	}

	switch runtime.GOOS {
	case "windows":
		user = os.Getenv("USERNAME")
	default:
		user = os.Getenv("USER")
	}
	if user == "" {
		// Make the error if it does not match any pattern
		return "", fmt.Errorf("failed to detect username. You can set ghq.user to your gitconfig")
	}
	return user, nil
}

func fillUsernameToPath(path string, forceMe bool) (string, error) {
	if !forceMe {
		completeUser, err := gitconfig.Bool("ghq.completeUser")
		if err != nil && !gitconfig.IsNotFound(err) {
			return path, err
		}
		if err == nil && !completeUser {
			return path + "/" + path, nil
		}
	}
	user, err := detectUserName()
	if err != nil {
		return path, err
	}
	return user + "/" + path, nil
}
