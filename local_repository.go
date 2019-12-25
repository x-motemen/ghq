package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Songmu/gitconfig"
	"github.com/motemen/ghq/logger"
	"github.com/saracen/walker"
)

const envGhqRoot = "GHQ_ROOT"

// LocalRepository represents local repository
type LocalRepository struct {
	FullPath  string
	RelPath   string
	RootPath  string
	PathParts []string

	repoPath   string
	vcsBackend *VCSBackend
}

// RepoPath returns local repository path
func (repo *LocalRepository) RepoPath() string {
	if repo.repoPath != "" {
		return repo.repoPath
	}
	return repo.FullPath
}

// LocalRepositoryFromFullPath resolve LocalRepository from file path
func LocalRepositoryFromFullPath(fullPath string, backend *VCSBackend) (*LocalRepository, error) {
	var relPath string

	roots, err := localRepositoryRoots(true)
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

// LocalRepositoryFromURL resolve LocalRepository from URL
func LocalRepositoryFromURL(remoteURL *url.URL) (*LocalRepository, error) {
	pathParts := append(
		[]string{remoteURL.Hostname()}, strings.Split(remoteURL.Path, "/")...,
	)
	relPath := strings.TrimSuffix(filepath.Join(pathParts...), ".git")
	pathParts[len(pathParts)-1] = strings.TrimSuffix(pathParts[len(pathParts)-1], ".git")

	var (
		localRepository *LocalRepository
		mu              sync.Mutex
	)
	// Find existing local repository first
	if err := walkAllLocalRepositories(func(repo *LocalRepository) {
		if repo.RelPath == relPath {
			mu.Lock()
			localRepository = repo
			mu.Unlock()
		}
	}); err != nil {
		return nil, err
	}

	if localRepository != nil {
		return localRepository, nil
	}

	prim, err := getRoot(remoteURL.String())
	if err != nil {
		return nil, err
	}

	// No local repository found, returning new one
	return &LocalRepository{
		FullPath:  filepath.Join(prim, relPath),
		RelPath:   relPath,
		RootPath:  prim,
		PathParts: pathParts,
	}, nil
}

func getRoot(u string) (string, error) {
	prim, err := gitconfig.Do("--path", "--get-urlmatch", "ghq.root", u)
	if err != nil && !gitconfig.IsNotFound(err) {
		return "", err
	}
	if prim == "" {
		prim, err = primaryLocalRepositoryRoot()
		if err != nil {
			return "", err
		}
	}
	return prim, nil
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

// NonHostPath returns non host path
func (repo *LocalRepository) NonHostPath() string {
	return strings.Join(repo.PathParts[1:], "/")
}

// list as bellow
// - "$GHQ_ROOT/github.com/motemen/ghq/cmdutil" // repo.FullPath
// - "$GHQ_ROOT/github.com/motemen/ghq"
// - "$GHQ_ROOT/github.com/motemen
func (repo *LocalRepository) repoRootCandidates() []string {
	hostRoot := filepath.Join(repo.RootPath, repo.PathParts[0])
	nonHostParts := repo.PathParts[1:]
	candidates := make([]string, len(nonHostParts))
	for i := 0; i < len(nonHostParts); i++ {
		candidates[i] = filepath.Join(append(
			[]string{hostRoot}, nonHostParts[0:len(nonHostParts)-i]...)...)
	}
	return candidates
}

// IsUnderPrimaryRoot or not
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

// VCS returns VCSBackend of the repository
func (repo *LocalRepository) VCS() (*VCSBackend, string) {
	if repo.vcsBackend == nil {
		for _, dir := range repo.repoRootCandidates() {
			backend := findVCSBackend(dir, "")
			if backend != nil {
				repo.vcsBackend = backend
				repo.repoPath = dir
				break
			}
		}
	}
	return repo.vcsBackend, repo.RepoPath()
}

var vcsContentsMap = map[string]*VCSBackend{
	".git":           GitBackend,
	".hg":            MercurialBackend,
	".svn":           SubversionBackend,
	"_darcs":         DarcsBackend,
	".bzr":           BazaarBackend,
	".fslckout":      FossilBackend, // file
	"_FOSSIL_":       FossilBackend, // file
	"CVS/Repository": cvsDummyBackend,
}

var vcsContents = [...]string{
	".git",
	".hg",
	".svn",
	"_darcs",
	".bzr",
	".fslckout",
	"._FOSSIL_",
	"CVS/Repository",
}

func findVCSBackend(fpath, vcs string) *VCSBackend {
	// When vcs is not empty, search only specified contents of vcs
	if vcs != "" {
		vcsBackend, ok := vcsRegistry[vcs]
		if !ok {
			return nil
		}
		for _, d := range vcsBackend.Contents {
			if _, err := os.Stat(filepath.Join(fpath, d)); err == nil {
				return vcsBackend
			}
		}
		return nil
	}
	for _, d := range vcsContents {
		if _, err := os.Stat(filepath.Join(fpath, d)); err == nil {
			return vcsContentsMap[d]
		}
	}
	return nil
}

func walkAllLocalRepositories(callback func(*LocalRepository)) error {
	return walkLocalRepositories("", callback)
}

func walkLocalRepositories(vcs string, callback func(*LocalRepository)) error {
	roots, err := localRepositoryRoots(true)
	if err != nil {
		return err
	}

	walkFn := func(fpath string, fi os.FileInfo) error {
		isSymlink := false
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			isSymlink = true
			realpath, err := filepath.EvalSymlinks(fpath)
			if err != nil {
				return nil
			}
			fi, err = os.Stat(realpath)
			if err != nil {
				return nil
			}
		}
		if !fi.IsDir() {
			return nil
		}
		vcsBackend := findVCSBackend(fpath, vcs)
		if vcsBackend == nil {
			return nil
		}

		repo, err := LocalRepositoryFromFullPath(fpath, vcsBackend)
		if err != nil || repo == nil {
			return nil
		}
		callback(repo)

		if isSymlink {
			return nil
		}
		return filepath.SkipDir
	}

	errCb := walker.WithErrorCallback(func(pathname string, err error) error {
		if os.IsPermission(errors.Unwrap(err)) {
			logger.Log("warning", fmt.Sprintf("%s: Permission denied", pathname))
			return nil
		}
		return err
	})

	for _, root := range roots {
		fi, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
		}
		if fi.Mode()&0444 == 0 {
			logger.Log("warning", fmt.Sprintf("%s: Permission denied", root))
			continue
		}
		if err := walker.Walk(root, walkFn, errCb); err != nil {
			return err
		}
	}
	return nil
}

var _home string

func getHome() (string, error) {
	if _home != "" {
		return _home, nil
	}
	var err error
	_home, err = os.UserHomeDir()
	return _home, err
}

var _localRepositoryRoots []string

// localRepositoryRoots returns locally cloned repositories' root directories.
// The root dirs are determined as following:
//
//   - If GHQ_ROOT environment variable is nonempty, use it as the only root dir.
//   - Otherwise, use the result of `git config --get-all ghq.root` as the dirs.
//   - Otherwise, fallback to the default root, `~/.ghq`.
func localRepositoryRoots(all bool) ([]string, error) {
	if len(_localRepositoryRoots) != 0 {
		return _localRepositoryRoots, nil
	}

	envRoot := os.Getenv(envGhqRoot)
	if envRoot != "" {
		_localRepositoryRoots = filepath.SplitList(envRoot)
	} else {
		var err error
		_localRepositoryRoots, err = gitconfig.PathAll("ghq.root")
		if err != nil && !gitconfig.IsNotFound(err) {
			return nil, err
		}
	}

	if len(_localRepositoryRoots) == 0 {
		homeDir, err := getHome()
		if err != nil {
			return nil, err
		}
		_localRepositoryRoots = []string{filepath.Join(homeDir, ".ghq")}
	}

	if all && envRoot == "" {
		roots, err := urlMatchLocalRepositoryRoots()
		if err != nil {
			return nil, err
		}
		_localRepositoryRoots = append(_localRepositoryRoots, roots...)
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

func urlMatchLocalRepositoryRoots() ([]string, error) {
	out, err := gitconfig.Do("--path", "--get-regexp", `^ghq\..+\.root$`)
	if err != nil {
		if gitconfig.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	items := strings.Split(out, "\x00")
	ret := make([]string, len(items))
	for i, kvStr := range items {
		kv := strings.SplitN(kvStr, "\n", 2)
		ret[i] = kv[1]
	}
	return ret, nil
}

func primaryLocalRepositoryRoot() (string, error) {
	roots, err := localRepositoryRoots(false)
	if err != nil {
		return "", err
	}
	return roots[0], nil
}
