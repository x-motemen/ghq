package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/motemen/ghq/cmdutil"
	"github.com/motemen/ghq/gitutil"
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
	pathComponents := strings.Split(strings.Trim(repo.url.Path, "/"), "/")
	return len(pathComponents) >= 2
}

func (repo *GitHubRepository) VCS() (*VCSBackend, *url.URL) {
	u, _ := url.Parse(repo.URL().String()) // clone
	pathComponents := strings.Split(strings.Trim(strings.TrimSuffix(u.Path, ".git"), "/"), "/")
	path := "/" + strings.Join(pathComponents[0:2], "/")
	if strings.HasSuffix(u.String(), ".git") {
		path += ".git"
	}
	u.Path = path
	return GitBackend, u
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
	if err := gitutil.HasFeatureConfigURLMatch(); err != nil {
		logger.Log("warning", err.Error())
	} else {
		// Respect 'ghq.url.https://ghe.example.com/.vcs' config variable
		// (in gitconfig:)
		//     [ghq "https://ghe.example.com/"]
		//     vcs = github
		vcs, err := gitutil.Config("--get-urlmatch", "ghq.vcs", repo.URL().String())
		if err != nil {
			logger.Log("error", err.Error())
		}
		if backend, ok := vcsRegistry[vcs]; ok {
			return backend, repo.URL()
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

	if url.Host == "hub.darcs.net" {
		return &DarksHubRepository{url}, nil
	}

	gheHosts, err := gitutil.ConfigAll("ghq.ghe.host")

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
