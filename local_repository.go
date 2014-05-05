package main

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/motemen/ghq/utils"
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
	if len(pathParts) != 3 { // host, user, project
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

// List of tail parts of relative path from the root directory (shortest first)
// for example, {"ghq", "motemen/ghq", "github.com/motemen/ghq"} for $root/github.com/motemen/ghq.
func (repo *LocalRepository) Subpaths() []string {
	tails := make([]string, len(repo.PathParts))

	for i, _ := range repo.PathParts {
		tails[i] = strings.Join(repo.PathParts[len(repo.PathParts)-(i+1):], "/")
	}

	return tails
}

func (repo *LocalRepository) NonHostPath() string {
	return strings.Join(repo.PathParts[1:], "/")
}

// Checks if any subpath of the local repository equals the query.
func (repo *LocalRepository) Matches(pathQuery string) bool {
	for _, p := range repo.Subpaths() {
		if p == pathQuery {
			return true
		}
	}

	return false
}

// TODO return err
func (repo *LocalRepository) VCS() *VCSBackend {
	var (
		fi  os.FileInfo
		err error
	)

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".git"))
	if err == nil && fi.IsDir() {
		return GitBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".hg"))
	if err == nil && fi.IsDir() {
		return MercurialBackend
	}

	return nil
}

func walkLocalRepositories(callback func(*LocalRepository)) {
	filepath.Walk(localRepositoriesRoot(), func(path string, fileInfo os.FileInfo, err error) error {
		repo, err := LocalRepositoryFromFullPath(path)
		if err != nil {
			return nil
		}

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

// Returns local cloned repositories' root.
// Uses the value of `git config ghq.root` or defaults to ~/.ghq.
func localRepositoriesRoot() string {
	if _localRepositoriesRoot != "" {
		return _localRepositoriesRoot
	}

	var err error
	_localRepositoriesRoot, err = GitConfig("ghq.root")
	utils.PanicIf(err)

	if _localRepositoriesRoot == "" {
		usr, err := user.Current()
		utils.PanicIf(err)

		_localRepositoriesRoot = path.Join(usr.HomeDir, ".ghq")
	}

	return _localRepositoriesRoot
}
