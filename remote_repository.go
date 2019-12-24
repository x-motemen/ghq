package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Songmu/gitconfig"
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

// URL reutrns URL of the repository
func (repo *GitHubRepository) URL() *url.URL {
	return repo.url
}

// IsValid determine if the repository is valid or not
func (repo *GitHubRepository) IsValid() bool {
	if strings.HasPrefix(repo.url.Path, "/blog/") {
		return false
	}
	pathComponents := strings.Split(strings.Trim(repo.url.Path, "/"), "/")
	return len(pathComponents) >= 2
}

// VCS returns VCSBackend of the repository
func (repo *GitHubRepository) VCS() (*VCSBackend, *url.URL) {
	origU := repo.URL()
	u := &url.URL{ // clone
		Scheme:   origU.Scheme,
		User:     origU.User,
		Host:     origU.Host,
		Path:     origU.Path,
		RawQuery: origU.RawQuery,
	}
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

// URL returns URL of the GistRepositroy
func (repo *GitHubGistRepository) URL() *url.URL {
	return repo.url
}

// IsValid determin if the gist rpository is valid or not
func (repo *GitHubGistRepository) IsValid() bool {
	return true
}

// VCS returns VCSBackend of the gist
func (repo *GitHubGistRepository) VCS() (*VCSBackend, *url.URL) {
	return GitBackend, repo.URL()
}

// DarksHubRepository represents DarcsHub Repository
type DarksHubRepository struct {
	url *url.URL
}

// URL returns URL of darks repository
func (repo *DarksHubRepository) URL() *url.URL {
	return repo.url
}

// IsValid determine if the darcshub repositroy is valid or not
func (repo *DarksHubRepository) IsValid() bool {
	return strings.Count(repo.url.Path, "/") == 2
}

// VCS returns VCSBackend of the DarcsHub repository
func (repo *DarksHubRepository) VCS() (*VCSBackend, *url.URL) {
	return DarcsBackend, repo.URL()
}

// OtherRepository represents other repository
type OtherRepository struct {
	url *url.URL
}

// URL returns URL of the repository
func (repo *OtherRepository) URL() *url.URL {
	return repo.url
}

// IsValid determine if the repository is valid or not
func (repo *OtherRepository) IsValid() bool {
	return true
}

// VCS detects VCSBackend of the OtherRepository
func (repo *OtherRepository) VCS() (*VCSBackend, *url.URL) {
	// Respect 'ghq.url.https://ghe.example.com/.vcs' config variable
	// (in gitconfig:)
	//     [ghq "https://ghe.example.com/"]
	//     vcs = github
	vcs, err := gitconfig.Do("--path", "--get-urlmatch", "ghq.vcs", repo.URL().String())
	if err != nil && !gitconfig.IsNotFound(err) {
		logger.Log("error", err.Error())
	}
	if backend, ok := vcsRegistry[vcs]; ok {
		return backend, repo.URL()
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

// NewRemoteRepository returns new RemoteRepository object from URL
func NewRemoteRepository(u *url.URL) (RemoteRepository, error) {
	repo := func() RemoteRepository {
		switch u.Host {
		case "github.com":
			return &GitHubRepository{u}
		case "gist.github.com":
			return &GitHubGistRepository{u}
		case "hub.darcs.net":
			return &DarksHubRepository{u}
		default:
			return &OtherRepository{u}
		}
	}()
	if !repo.IsValid() {
		return nil, fmt.Errorf("Not a valid repository: %s", u)
	}
	return repo, nil
}
