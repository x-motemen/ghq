package main

import (
	"fmt"
	"net/url"
	"os"
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
	var relPath string

	roots, err := localRepositoryRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		if !strings.HasPrefix(fullPath, root) {
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

	return &LocalRepository{
		FullPath:  fullPath,
		RelPath:   filepath.ToSlash(relPath),
		PathParts: pathParts}, nil
}

func LocalRepositoryFromURL(remoteURL *url.URL) (*LocalRepository, error) {
	pathParts := append(
		[]string{remoteURL.Host}, strings.Split(remoteURL.Path, "/")...,
	)
	relPath := strings.TrimSuffix(path.Join(pathParts...), ".git")

	var localRepository *LocalRepository

	// Find existing local repository first
	if err := walkLocalRepositories(func(repo *LocalRepository) {
		if repo.RelPath == relPath {
			localRepository = repo
		}
	}); err != nil {
		return nil, err
	}

	if localRepository != nil {
		return localRepository, nil
	}

	prim, err := primaryLocalRepositoryRoot()
	if err != nil {
		return nil, err
	}
	// No local repository found, returning new one
	return &LocalRepository{
		FullPath:  path.Join(prim, relPath),
		RelPath:   relPath,
		PathParts: pathParts,
	}, nil
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
	prim, err := primaryLocalRepositoryRoot()
	if err != nil {
		return false
	}
	return strings.HasPrefix(repo.FullPath, prim)
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

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".fslckout"))
	if err == nil && fi.IsDir() {
		return FossilBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, "_FOSSIL_"))
	if err == nil && fi.IsDir() {
		return FossilBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, "CVS"))
	if err == nil && fi.IsDir() {
		return cvsDummyBackend
	}

	fi, err = os.Stat(filepath.Join(repo.FullPath, ".bzr"))
	if err == nil && fi.IsDir() {
		return BazaarBackend
	}


	return nil
}

var vcsDirs = []string{".git", ".svn", ".hg", "_darcs", ".fslckout", "_FOSSIL_", "CVS", ".bzr"}

func walkLocalRepositories(callback func(*LocalRepository)) error {
	roots, err := localRepositoryRoots()
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := filepath.Walk(root, func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil || fileInfo == nil {
				return nil
			}

			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				realpath, err := filepath.EvalSymlinks(path)
				if err != nil {
					return nil
				}
				fileInfo, err = os.Stat(realpath)
				if err != nil {
					return nil
				}
			}
			if !fileInfo.IsDir() {
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
		}); err != nil {
			return err
		}
	}
	return nil
}

var _localRepositoryRoots []string

// localRepositoryRoots returns locally cloned repositories' root directories.
// The root dirs are determined as following:
//
//   - If GHQ_ROOT environment variable is nonempty, use it as the only root dir.
//   - Otherwise, use the result of `git config --get-all ghq.root` as the dirs.
//   - Otherwise, fallback to the default root, `~/.ghq`.
func localRepositoryRoots() ([]string, error) {
	if len(_localRepositoryRoots) != 0 {
		return _localRepositoryRoots, nil
	}

	envRoot := os.Getenv("GHQ_ROOT")
	if envRoot != "" {
		_localRepositoryRoots = filepath.SplitList(envRoot)
	} else {
		var err error
		if _localRepositoryRoots, err = GitConfigAll("ghq.root"); err != nil {
			return nil, err
		}
	}

	if len(_localRepositoryRoots) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		_localRepositoryRoots = []string{filepath.Join(homeDir, ".ghq")}
	}

	for i, v := range _localRepositoryRoots {
		path := filepath.Clean(v)
		if _, err := os.Stat(path); err == nil {
			if path, err = filepath.EvalSymlinks(path); err != nil {
				return nil, err
			}
		}
		if !filepath.IsAbs(path) {
			var err error
			if path, err = filepath.Abs(path); err != nil {
				return nil, err
			}
		}
		_localRepositoryRoots[i] = path
	}

	return _localRepositoryRoots, nil
}

func primaryLocalRepositoryRoot() (string, error) {
	roots, err := localRepositoryRoots()
	if err != nil {
		return "", err
	}
	return roots[0], nil
}
