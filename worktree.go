package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// isNotADirectory returns true if err indicates a "not a directory" condition
// (e.g., trying to traverse a path component that is a regular file).
func isNotADirectory(err error) bool {
	return errors.Is(err, syscall.ENOTDIR)
}

// isLinkedGitDir checks whether dir has a .git file (not directory) with a
// gitdir: reference. This is the case for both linked worktrees and
// submodules — either way, the directory cannot be migrated independently.
// When true, it returns the resolved gitdir target path.
func isLinkedGitDir(dir string) (bool, string, error) {
	dotGit := filepath.Join(dir, ".git")
	fi, err := os.Lstat(dotGit)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", err
	}

	// .git is a directory → regular repo, safe to migrate
	if fi.IsDir() {
		return false, "", nil
	}

	// .git is a file → linked checkout (worktree or submodule)
	content, err := os.ReadFile(dotGit)
	if err != nil {
		return false, "", err
	}

	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		return false, "", nil
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")

	// Resolve relative paths
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(dir, gitdir)
	}
	gitdir = filepath.Clean(gitdir)

	return true, gitdir, nil
}

// isWorktreeGitDir returns true if gitdirTarget looks like a worktree entry
// (.git/worktrees/<name>) rather than a submodule (.git/modules/<name>).
func isWorktreeGitDir(gitdirTarget string) bool {
	return strings.Contains(filepath.ToSlash(gitdirTarget), ".git/worktrees/")
}

// hasLinkedWorktrees reports whether the Git repository at dir has any linked
// worktrees (entries under .git/worktrees/).
//
// Known limitation: bare repos store worktrees in <bare-repo>/worktrees/
// (no .git/ prefix). This check only looks at .git/worktrees/ and would
// miss bare repo worktrees.
func hasLinkedWorktrees(dir string) (bool, error) {
	worktreesDir := filepath.Join(dir, ".git", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) || isNotADirectory(err) {
			return false, nil
		}
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			return true, nil
		}
	}
	return false, nil
}

// listLinkedWorktreePaths reads .git/worktrees/*/gitdir in dir and returns
// the worktree working-directory paths.
func listLinkedWorktreePaths(dir string) ([]string, error) {
	worktreesDir := filepath.Join(dir, ".git", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) || isNotADirectory(err) {
			return nil, nil
		}
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gitdirFile := filepath.Join(worktreesDir, e.Name(), "gitdir")
		content, err := os.ReadFile(gitdirFile)
		if err != nil {
			continue
		}
		wtPath := strings.TrimSpace(string(content))
		if wtPath == "" {
			continue
		}

		// Resolve to native path
		wtPath = filepath.FromSlash(wtPath)
		if !filepath.IsAbs(wtPath) {
			wtPath = filepath.Join(worktreesDir, e.Name(), wtPath)
		}
		wtPath = filepath.Clean(wtPath)

		// The gitdir file stores the path to the worktree's .git file
		// (e.g., "/path/to/wt/.git"); strip trailing /.git to get working dir.
		wtDir := strings.TrimSuffix(filepath.ToSlash(wtPath), "/.git")
		paths = append(paths, filepath.FromSlash(wtDir))
	}
	return paths, nil
}

// resolveMainRepoDir resolves the main repository working directory from a
// worktree's gitdir target path (e.g., /path/to/main/.git/worktrees/<name>).
// It reads the commondir file to find the shared .git directory.
func resolveMainRepoDir(gitdirTarget string) (string, error) {
	commondirFile := filepath.Join(gitdirTarget, "commondir")
	content, err := os.ReadFile(commondirFile)
	if err != nil {
		return "", fmt.Errorf("failed to read commondir: %w", err)
	}
	commondir := strings.TrimSpace(string(content))
	if !filepath.IsAbs(commondir) {
		commondir = filepath.Join(gitdirTarget, commondir)
	}
	commondir = filepath.Clean(commondir)
	// commondir points to the .git directory; the working tree is its parent
	return filepath.Dir(commondir), nil
}

// repairWorktreeBackPointers reads .git/worktrees/*/gitdir in destDir and
// returns the current worktree working-directory paths. For worktrees that
// were inside the old repo directory (oldDir), it rewrites the gitdir file
// to reflect the new location so that a subsequent "git worktree repair"
// can match them.
func repairWorktreeBackPointers(oldDir, destDir string) ([]string, error) {
	worktreesDir := filepath.Join(destDir, ".git", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Normalize paths to forward slashes for comparison, since Git uses forward slashes
	// in gitdir files even on Windows
	oldPrefixNorm := filepath.ToSlash(oldDir) + "/"
	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gitdirFile := filepath.Join(worktreesDir, e.Name(), "gitdir")
		content, err := os.ReadFile(gitdirFile)
		if err != nil {
			continue // skip entries without a gitdir file
		}
		wtPath := strings.TrimSpace(string(content))
		if wtPath == "" {
			continue
		}

		// Internal worktree: moved along with the repo → fix back-pointer
		// Normalize wtPath for comparison since Git writes forward slashes on all platforms
		wtPathNorm := filepath.ToSlash(wtPath)
		if strings.HasPrefix(wtPathNorm, oldPrefixNorm) {
			// Compute new path: take the relative portion and join with destDir
			relativePart := wtPathNorm[len(oldPrefixNorm):]
			newPath := filepath.ToSlash(filepath.Join(destDir, relativePart))
			if err := os.WriteFile(gitdirFile, []byte(newPath+"\n"), 0644); err != nil {
				return nil, fmt.Errorf("failed to rewrite gitdir for worktree %s: %w", e.Name(), err)
			}
			wtPath = newPath
		}

		// The gitdir file stores the path to the worktree's .git file
		// (e.g., "/path/to/wt/.git"), but git worktree repair expects
		// the worktree working directory (e.g., "/path/to/wt").
		// Since wtPath is normalized to forward slashes, trim "/.git"
		wtDir := strings.TrimSuffix(wtPath, "/.git")
		paths = append(paths, wtDir)
	}
	return paths, nil
}
