package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/motemen/ghq/utils"
)

// A RemoteRepository represents a remote repository.
type RemoteRepository interface {
	// The repository URL.
	URL() *url.URL
	// Checks if the URL is valid.
	IsValid() bool
	// The VCS backend that hosts the repository.
	VCS() *VCSBackend
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

func (repo *GitHubRepository) VCS() *VCSBackend {
	return GitBackend
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

func (repo *GitHubGistRepository) VCS() *VCSBackend {
	return GitBackend
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

func (repo *GoogleCodeRepository) VCS() *VCSBackend {
	if utils.RunSilently("hg", "identify", repo.url.String()) == nil {
		return MercurialBackend
	} else if utils.RunSilently("git", "ls-remote", repo.url.String()) == nil {
		return GitBackend
	} else {
		return nil
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

func (repo *DarksHubRepository) VCS() *VCSBackend {
	return DarcsBackend
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

func (repo *BluemixRepository) VCS() *VCSBackend {
	return GitBackend
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

func (repo *OtherRepository) VCS() *VCSBackend {
	if GitHasFeatureConfigURLMatch() {
		// Respect 'ghq.url.https://ghe.example.com/.vcs' config variable
		// (in gitconfig:)
		//     [ghq "https://ghe.example.com/"]
		//     vcs = github
		vcs, err := GitConfig("--get-urlmatch", "ghq.vcs", repo.URL().String())
		if err != nil {
			utils.Log("error", err.Error())
		}

		if vcs == "git" || vcs == "github" {
			return GitBackend
		}

		if vcs == "svn" || vcs == "subversion" {
			return SubversionBackend
		}

		if vcs == "git-svn" {
			return GitsvnBackend
		}

		if vcs == "hg" || vcs == "mercurial" {
			return MercurialBackend
		}

		if vcs == "darcs" {
			return DarcsBackend
		}
	} else {
		utils.Log("warning", "This version of Git does not support `config --get-urlmatch`; per-URL settings are not available")
	}

	// Detect VCS backend automatically
	if utils.RunSilently("git", "ls-remote", repo.url.String()) == nil {
		return GitBackend
	} else if utils.RunSilently("hg", "identify", repo.url.String()) == nil {
		return MercurialBackend
	} else if utils.RunSilently("svn", "info", repo.url.String()) == nil {
		return SubversionBackend
	} else {
		return nil
	}
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
