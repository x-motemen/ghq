package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/mitchellh/go-homedir"
	"github.com/motemen/ghq/utils"
)

type LocalRepository struct {
	FullPath  string
	RelPath   string
	PathParts []string
}

func LocalRepositoryFromFullPath(fullPath string) (*LocalRepository, error) {
	var relPath string

	for _, root := range localRepositoryRoots() {
		if strings.HasPrefix(fullPath, root) == false {
			continue
		}

		var err error
		relPath, err = filepath.Rel(root, fullPath)
		if err == nil {
			break
		}
	}

	if relPath == "" {
		return nil, fmt.Errorf("no local repository found for: %s", fullPath)
	}

	pathParts := strings.Split(relPath, string(filepath.Separator))

	return &LocalRepository{fullPath, filepath.ToSlash(relPath), pathParts}, nil
}

func LocalRepositoryFromURL(remoteURL *url.URL) *LocalRepository {
	pathParts := append(
		[]string{remoteURL.Host}, strings.Split(remoteURL.Path, "/")...,
	)
	relPath := strings.TrimSuffix(path.Join(pathParts...), ".git")

	var localRepository *LocalRepository

	// Find existing local repository first
	walkLocalRepositories(func(repo *LocalRepository) {
		if repo.RelPath == relPath {
			localRepository = repo
		}
	})

	if localRepository != nil {
		return localRepository
	}

	// No local repository found, returning new one
	return &LocalRepository{
		path.Join(primaryLocalRepositoryRoot(), relPath),
		relPath,
		pathParts,
	}
}

// Subpaths returns lists of tail parts of relative path from the root directory (shortest first)
// for example, {"ghq", "motemen/ghq", "github.com/motemen/ghq"} for $root/github.com/motemen/ghq.
func (repo *LocalRepository) Subpaths() []string {
	tails := make([]string, len(repo.PathParts))

	for i := range repo.PathParts {
		tails[i] = strings.Join(repo.PathParts[len(repo.PathParts)-(i+1):], "/")
	}

	return tails
}

func (repo *LocalRepository) NonHostPath() string {
	return strings.Join(repo.PathParts[1:], "/")
}

func (repo *LocalRepository) IsUnderPrimaryRoot() bool {
	return strings.HasPrefix(repo.FullPath, primaryLocalRepositoryRoot())
}

// Matches checks if any subpath of the local repository equals the query.
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

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".git/svn"))
	if err == nil && fi.IsDir() {
		return GitsvnBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".git"))
	if err == nil && fi.IsDir() {
		return GitBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".svn"))
	if err == nil && fi.IsDir() {
		return SubversionBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".hg"))
	if err == nil && fi.IsDir() {
		return MercurialBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, "_darcs"))
	if err == nil && fi.IsDir() {
		return DarcsBackend
	}

	return nil
}

var vcsDirs = []string{".git", ".svn", ".hg", "_darcs"}

func walkLocalRepositories(callback func(*LocalRepository)) {
	for _, root := range localRepositoryRoots() {
		godirwalk.Walk(root, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
				if de.IsDir() == false {
					return nil
				}

				vcsDirFound := false
				for _, d := range vcsDirs {
					_, err := os.Stat(filepath.Join(path, d))
					if err == nil {
						vcsDirFound = true
						break
					}
				}

				if !vcsDirFound {
					return nil
				}

				repo, err := LocalRepositoryFromFullPath(path)
				if err != nil {
					return nil
				}

				if repo == nil {
					return nil
				}
				callback(repo)
				return filepath.SkipDir
			},
		})
	}
}

var _localRepositoryRoots []string

// localRepositoryRoots returns locally cloned repositories' root directories.
// The root dirs are determined as following:
//
//   - If GHQ_ROOT environment variable is nonempty, use it as the only root dir.
//   - Otherwise, use the result of `git config --get-all ghq.root` as the dirs.
//   - Otherwise, fallback to the default root, `~/.ghq`.
//
// TODO: More fancy default directory path?
func localRepositoryRoots() []string {
	if len(_localRepositoryRoots) != 0 {
		return _localRepositoryRoots
	}

	envRoot := os.Getenv("GHQ_ROOT")
	if envRoot != "" {
		_localRepositoryRoots = filepath.SplitList(envRoot)
	} else {
		var err error
		_localRepositoryRoots, err = GitConfigAll("ghq.root")
		utils.PanicIf(err)
	}

	if len(_localRepositoryRoots) == 0 {
		homeDir, err := homedir.Dir()
		utils.PanicIf(err)

		_localRepositoryRoots = []string{filepath.Join(homeDir, ".ghq")}
	}

	for i, v := range _localRepositoryRoots {
		path := filepath.Clean(v)
		if _, err := os.Stat(path); err == nil {
			_localRepositoryRoots[i], err = filepath.EvalSymlinks(path)
			utils.PanicIf(err)
		} else {
			_localRepositoryRoots[i] = path
		}
	}

	return _localRepositoryRoots
}

func primaryLocalRepositoryRoot() string {
	return localRepositoryRoots()[0]
}
