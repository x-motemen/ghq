package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type LocalRepository struct {
	FullPath  string
	RelPath   string
	RootPath  string
	PathParts []string

	vcsBackend *VCSBackend
}

func LocalRepositoryFromFullPath(fullPath string, backend *VCSBackend) (*LocalRepository, error) {
	var relPath string

	roots, err := localRepositoryRoots()
	if err != nil {
		return nil, err
	}
	var root string
	for _, root = range roots {
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
		FullPath:   fullPath,
		RelPath:    filepath.ToSlash(relPath),
		RootPath:   root,
		PathParts:  pathParts,
		vcsBackend: backend,
	}, nil
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
		RootPath:  prim,
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

func (repo *LocalRepository) VCS() *VCSBackend {
	if repo.vcsBackend == nil {
		repo.vcsBackend = findVCSBackend(repo.FullPath)
	}
	return repo.vcsBackend
}

var vcsContentsMap = map[string]*VCSBackend{
	".git/svn":       GitsvnBackend,
	".git":           GitBackend,
	".svn":           SubversionBackend,
	".hg":            MercurialBackend,
	"_darcs":         DarcsBackend,
	".fslckout":      FossilBackend, // file
	"_FOSSIL_":       FossilBackend, // file
	"CVS/Repository": cvsDummyBackend,
	".bzr":           BazaarBackend,
}

var vcsContents = make([]string, 0, len(vcsContentsMap))

func init() {
	for k := range vcsContentsMap {
		vcsContents = append(vcsContents, k)
	}
	// Sort in order of length.
	// This is to check git/svn before git.
	sort.Slice(vcsContents, func(i, j int) bool {
		return len(vcsContents[i]) > len(vcsContents[j])
	})
}

func findVCSBackend(fpath string) *VCSBackend {
	for _, d := range vcsContents {
		if _, err := os.Stat(filepath.Join(fpath, d)); err == nil {
			return vcsContentsMap[d]
		}
	}
	return nil
}

func walkLocalRepositories(callback func(*LocalRepository)) error {
	roots, err := localRepositoryRoots()
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := filepath.Walk(root, func(fpath string, fileInfo os.FileInfo, err error) error {
			if err != nil || fileInfo == nil {
				return nil
			}

			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				realpath, err := filepath.EvalSymlinks(fpath)
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
			vcsBackend := findVCSBackend(fpath)
			if vcsBackend == nil {
				return nil
			}

			repo, err := LocalRepositoryFromFullPath(fpath, vcsBackend)
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
