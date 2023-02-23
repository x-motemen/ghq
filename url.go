package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Songmu/gitconfig"
	"github.com/x-motemen/ghq/logger"
)

// Convert SCP-like URL to SSH URL(e.g. [user@]host.xz:path/to/repo.git/)
// ref. http://git-scm.com/docs/git-fetch#_git_urls
// (golang hasn't supported Perl-like negative look-behind match)
var (
	hasSchemePattern          = regexp.MustCompile("^[^:]+://")
	scpLikeURLPattern         = regexp.MustCompile("^([^@]+@)?([^:]+):(/?.+)$")
	looksLikeAuthorityPattern = regexp.MustCompile(`[A-Za-z0-9]\.[A-Za-z]+(?::\d{1,5})?$`)
	codecommitLikeURLPattern  = regexp.MustCompile(`^(codecommit):(?::([a-z][a-z0-9-]+):)?//(?:([^]]+)@)?([\w\.-]+)$`)
)

func newURL(ref string, ssh, forceMe bool) (*url.URL, error) {
	// If argURL is a "./foo" or "../bar" form,
	// find repository name trailing after github.com/USER/.
	ref = filepath.ToSlash(ref)
	parts := strings.Split(ref, "/")
	if parts[0] == "." || parts[0] == ".." {
		if wd, err := os.Getwd(); err == nil {
			path := filepath.Clean(filepath.Join(wd, filepath.Join(parts...)))

			var localRepoRoot string
			roots, err := localRepositoryRoots(true)
			if err != nil {
				return nil, err
			}
			for _, r := range roots {
				p := strings.TrimPrefix(path, r+string(filepath.Separator))
				if p != path && (localRepoRoot == "" || len(p) < len(localRepoRoot)) {
					localRepoRoot = filepath.ToSlash(p)
				}
			}

			if localRepoRoot != "" {
				// Guess it
				logger.Log("resolved", fmt.Sprintf("relative %q to %q", ref, "https://"+localRepoRoot))
				ref = "https://" + localRepoRoot
			}
		}
	}

	if codecommitLikeURLPattern.MatchString(ref) {
		// SEE ALSO:
		// https://github.com/aws/git-remote-codecommit/blob/master/git_remote_codecommit/__init__.py#L68
		matched := codecommitLikeURLPattern.FindStringSubmatch(ref)
		region := matched[2]

		if matched[2] == "" {
			// Region detection priority:
			// 1. Explicit specification (codecommit::region://...)
			// 2. Environment variables
			//     a. AWS_REGION (implicit priority)
			//     b. AWS_DEFAULT_REGION
			// 3. AWS CLI profiles
			// SEE ALSO:
			// https://docs.aws.amazon.com/ja_jp/cli/latest/userguide/cli-configure-quickstart.html#cli-configure-quickstart-precedence
			var exists bool
			region, exists = os.LookupEnv("AWS_REGION")
			if !exists {
				region, exists = os.LookupEnv("AWS_DEFAULT_REGION")
			}

			if !exists {
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				cmd := exec.Command("aws", "configure", "get", "region")
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				if err != nil {
					if stderr.String() == "" {
						fmt.Fprintln(os.Stderr, "You must specify a region. You can also configure your region by running \"aws configure\".")
					} else {
						fmt.Fprint(os.Stderr, stderr.String())
					}
					os.Exit(1)
				}

				region = strings.TrimSpace(stdout.String())
			}
		}

		return &url.URL{
			Scheme: matched[1],
			Host:   region,
			User:   url.User(matched[3]),
			Path:   matched[4],
			Opaque: ref,
		}, nil
	}

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

	u, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		if !strings.Contains(u.Path, "/") {
			u.Path, err = fillUsernameToPath(u.Path, forceMe)
			if err != nil {
				return nil, err
			}
		}
		u.Scheme = "https"
		u.Host = "github.com"
		if u.Path[0] != '/' {
			u.Path = "/" + u.Path
		}
	}

	if ssh {
		// Assume Git repository if `-p` is given.
		if u, err = convertGitURLHTTPToSSH(u); err != nil {
			return nil, fmt.Errorf("could not convert URL %q: %w", u, err)
		}
	}

	return u, nil
}

func convertGitURLHTTPToSSH(u *url.URL) (*url.URL, error) {
	user := "git"
	if u.User != nil {
		user = u.User.Username()
	}
	sshURL := fmt.Sprintf("ssh://%s@%s%s", user, u.Host, u.Path)
	return u.Parse(sshURL)
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
