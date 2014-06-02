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

func NewRemoteRepository(url *url.URL) (RemoteRepository, error) {
	if url.Host == "github.com" {
		return &GitHubRepository{url}, nil
	}

	if url.Host == "code.google.com" {
		return &GoogleCodeRepository{url}, nil
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

	return nil, fmt.Errorf("unsupported host: %s", url.Host)
}
