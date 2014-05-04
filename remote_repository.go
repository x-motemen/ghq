package main

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type RemoteRepository interface {
	RepositoryURL() *url.URL
	IsValid() bool
	VCS() *VCSBackend
}

type GitHubRepository struct {
	*url.URL
}

func (repo *GitHubRepository) RepositoryURL() *url.URL {
	return repo.URL
}

func (repo *GitHubRepository) IsValid() bool {
	if strings.HasPrefix(repo.Path, "/blog/") {
		return false
	}

	// must be /{user}/{project}
	if len(strings.Split(repo.Path, "/")) != 3 {
		return false
	}

	return true
}

func (repo *GitHubRepository) VCS() *VCSBackend {
	return GitBackend
}

type VCSBackend struct {
	Clone  func(*url.URL, string) error
	Update func(string) error
}

var GitBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return Git("clone", remote.String(), local)
	},
	Update: func(local string) error {
		err := os.Chdir(local)
		if err != nil {
			return err
		}

		return Git("remote", "update")
	},
}

func NewRemoteRepository(url *url.URL) (RemoteRepository, error) {
	if url.Host == "github.com" {
		return &GitHubRepository{url}, nil
	}

	return nil, errors.New("Unsupported domain")
}
