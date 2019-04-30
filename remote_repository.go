package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/motemen/ghq/cmdutil"
	"github.com/motemen/ghq/logger"
)

// A RemoteRepository represents a remote repository.
type RemoteRepository interface {
	// The repository URL.
	URL() *url.URL
	// Checks if the URL is valid.
	IsValid() bool
	// The VCS backend that hosts the repository.
	VCS() (*VCSBackend, *url.URL)
}

// A GitHubRepository represents a GitHub repository. Impliments RemoteRepository.
type GitHubRepository struct {
	url *url.URL
}

func (repo *GitHubRepository) URL() *url.URL {
	return repo.url
}

func (repo *GitHubRepository) IsValid() bool {
	if strings.HasPrefix(repo.url.Path, "/blog/") {
		return false
	}

	// must be /{user}/{project}/?
	pathComponents := strings.Split(strings.TrimRight(repo.url.Path, "/"), "/")
	if len(pathComponents) != 3 {
		return false
	}

	return true
}

func (repo *GitHubRepository) VCS() (*VCSBackend, *url.URL) {
	return GitBackend, repo.URL()
}

// A GitHubGistRepository represents a GitHub Gist repository.
type GitHubGistRepository struct {
	url *url.URL
}

func (repo *GitHubGistRepository) URL() *url.URL {
	return repo.url
}

func (repo *GitHubGistRepository) IsValid() bool {
	return true
}

func (repo *GitHubGistRepository) VCS() (*VCSBackend, *url.URL) {
	return GitBackend, repo.URL()
}

type GoogleCodeRepository struct {
	url *url.URL
}

func (repo *GoogleCodeRepository) URL() *url.URL {
	return repo.url
}

var validGoogleCodePathPattern = regexp.MustCompile(`^/p/[^/]+/?$`)

func (repo *GoogleCodeRepository) IsValid() bool {
	return validGoogleCodePathPattern.MatchString(repo.url.Path)
}

func (repo *GoogleCodeRepository) VCS() (*VCSBackend, *url.URL) {
	if cmdutil.RunSilently("hg", "identify", repo.url.String()) == nil {
		return MercurialBackend, repo.URL()
	} else if cmdutil.RunSilently("git", "ls-remote", repo.url.String()) == nil {
		return GitBackend, repo.URL()
	} else {
		return nil, nil
	}
}

type DarksHubRepository struct {
	url *url.URL
}

func (repo *DarksHubRepository) URL() *url.URL {
	return repo.url
}

func (repo *DarksHubRepository) IsValid() bool {
	return strings.Count(repo.url.Path, "/") == 2
}

func (repo *DarksHubRepository) VCS() (*VCSBackend, *url.URL) {
	return DarcsBackend, repo.URL()
}

type BluemixRepository struct {
	url *url.URL
}

func (repo *BluemixRepository) URL() *url.URL {
	return repo.url
}

var validBluemixPathPattern = regexp.MustCompile(`^/git/[^/]+/[^/]+$`)

func (repo *BluemixRepository) IsValid() bool {
	return validBluemixPathPattern.MatchString(repo.url.Path)
}

func (repo *BluemixRepository) VCS() (*VCSBackend, *url.URL) {
	return GitBackend, repo.URL()
}

type OtherRepository struct {
	url *url.URL
}

func (repo *OtherRepository) URL() *url.URL {
	return repo.url
}

func (repo *OtherRepository) IsValid() bool {
	return true
}

func (repo *OtherRepository) VCS() (*VCSBackend, *url.URL) {
	if err := GitHasFeatureConfigURLMatch(); err != nil {
		logger.Log("warning", err.Error())
	} else {
		// Respect 'ghq.url.https://ghe.example.com/.vcs' config variable
		// (in gitconfig:)
		//     [ghq "https://ghe.example.com/"]
		//     vcs = github
		vcs, err := GitConfig("--get-urlmatch", "ghq.vcs", repo.URL().String())
		if err != nil {
			logger.Log("error", err.Error())
		}

		if vcs == "git" || vcs == "github" {
			return GitBackend, repo.URL()
		}

		if vcs == "svn" || vcs == "subversion" {
			return SubversionBackend, repo.URL()
		}

		if vcs == "git-svn" {
			return GitsvnBackend, repo.URL()
		}

		if vcs == "hg" || vcs == "mercurial" {
			return MercurialBackend, repo.URL()
		}

		if vcs == "darcs" {
			return DarcsBackend, repo.URL()
		}

		if vcs == "fossil" {
			return FossilBackend, repo.URL()
		}

		if vcs == "bazaar" || vcs == "bzr" {
			return BazaarBackend, repo.URL()
		}
	}

	// Detect VCS backend automatically
	if cmdutil.RunSilently("git", "ls-remote", repo.url.String()) == nil {
		return GitBackend, repo.URL()
	}

	vcs, repoURL, err := detectGoImport(repo.url)
	if err == nil {
		// vcs == "mod" (modproxy) not supported yet
		return vcsRegistry[vcs], repoURL
	}

	if cmdutil.RunSilently("hg", "identify", repo.url.String()) == nil {
		return MercurialBackend, repo.URL()
	}

	if cmdutil.RunSilently("svn", "info", repo.url.String()) == nil {
		return SubversionBackend, repo.URL()
	}

	return nil, nil
}

func NewRemoteRepository(url *url.URL) (RemoteRepository, error) {
	if url.Host == "github.com" {
		return &GitHubRepository{url}, nil
	}

	if url.Host == "gist.github.com" {
		return &GitHubGistRepository{url}, nil
	}

	if url.Host == "code.google.com" {
		return &GoogleCodeRepository{url}, nil
	}

	if url.Host == "hub.darcs.net" {
		return &DarksHubRepository{url}, nil
	}

	if url.Host == "hub.jazz.net" {
		return &BluemixRepository{url}, nil
	}

	gheHosts, err := GitConfigAll("ghq.ghe.host")

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve GH:E hostname from .gitconfig: %s", err)
	}

	for _, host := range gheHosts {
		if url.Host == host {
			return &GitHubRepository{url}, nil
		}
	}

	return &OtherRepository{url}, nil
}
