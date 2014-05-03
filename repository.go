package main

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

type LocalRepository struct {
	FullPath  string
	RelPath   string
	PathParts []string
}

func LocalRepositoryFromFullPath(fullPath string) (*LocalRepository, error) {
	relPath, err := filepath.Rel(localRepositoriesRoot(), fullPath)
	if err != nil {
		return nil, err
	}

	pathParts := strings.Split(relPath, string(filepath.Separator))
	if len(pathParts) != 3 {
		return nil, nil
	}

	return &LocalRepository{fullPath, relPath, pathParts}, nil
}

func LocalRepositoryFromPathParts(pathParts []string) *LocalRepository {
	relPath := path.Join(pathParts...)
	return &LocalRepository{
		path.Join(localRepositoriesRoot(), relPath),
		relPath,
		pathParts,
	}
}

func LocalRepositoryFromGitHubURL(u *GitHubURL) *LocalRepository {
	return LocalRepositoryFromPathParts([]string{"github.com", u.User, u.Repo})
}

func (repo *LocalRepository) PathTails() []string {
	tails := make([]string, len(repo.PathParts))

	for i, _ := range repo.PathParts {
		tails[i] = strings.Join(repo.PathParts[i:], "/")
	}

	return tails
}

func (repo *LocalRepository) NonHostPath() string {
	return strings.Join(repo.PathParts[1:], "/")
}

func (repo *LocalRepository) Matches(pathQuery string) bool {
	for _, p := range repo.PathTails() {
		if p == pathQuery {
			return true
		}
	}

	return false
}

func walkLocalRepositories(callback func(*LocalRepository)) {
	filepath.Walk(localRepositoriesRoot(), func(path string, fileInfo os.FileInfo, err error) error {
		repo, err := LocalRepositoryFromFullPath(path)
		mustBeOkay(err)

		if repo != nil {
			callback(repo)
			return filepath.SkipDir
		} else {
			return nil
		}
	})

	return
}

var _localRepositoriesRoot string

func localRepositoriesRoot() string {
	if _localRepositoriesRoot != "" {
		return _localRepositoriesRoot
	}

	var err error
	_localRepositoriesRoot, err = GitConfig("ghq.root")
	mustBeOkay(err)

	if _localRepositoriesRoot == "" {
		usr, err := user.Current()
		mustBeOkay(err)

		_localRepositoriesRoot = path.Join(usr.HomeDir, ".ghq")
	}

	return _localRepositoriesRoot
}
