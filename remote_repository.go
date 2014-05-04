package main

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/motemen/ghq/utils"
)

// Represents a remote repository.
type RemoteRepository interface {
	// The repository URL.
	URL() *url.URL
	// Checks if the URL is valid.
	IsValid() bool
	// The VCS backend that hosts the repository.
	VCS() *VCSBackend
}

// Represents a GitHub repository. Impliments RemoteRepository.
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

	// must be /{user}/{project}
	if len(strings.Split(repo.url.Path, "/")) != 3 {
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

	return nil, errors.New("Unsupported host")
}
