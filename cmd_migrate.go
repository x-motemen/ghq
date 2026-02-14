package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/otiai10/copy"
	"github.com/urfave/cli/v2"
	"github.com/x-motemen/ghq/logger"
)

func doMigrate(c *cli.Context) error {
	var (
		repoDir     = c.Args().First()
		dry         = c.Bool("dry-run")
		skipConfirm = c.Bool("y")
		w           = c.App.Writer
	)

	if repoDir == "" {
		return fmt.Errorf("repository directory is required")
	}

	// Resolve directory (supports both absolute and relative paths)
	absDir, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	// Check if the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory %q does not exist", absDir)
	} else if err != nil {
		return fmt.Errorf("failed to access directory %q: %w", absDir, err)
	}

	// Detect VCS backend
	vcsBackend := findVCSBackend(absDir, "")
	if vcsBackend == nil {
		return fmt.Errorf("failed to detect VCS backend in %q", absDir)
	}

	// Refuse to migrate a linked Git checkout (worktree or submodule).
	// These have a .git file referencing a parent repo; moving them alone
	// breaks the link.
	if vcsBackend == GitBackend {
		if linked, target, err := isLinkedGitDir(absDir); err != nil {
			return fmt.Errorf("failed to check .git link status: %w", err)
		} else if linked {
			return fmt.Errorf("directory %q has a .git file linking to %q; it is a worktree or submodule and cannot be migrated independently", absDir, target)
		}
	}

	// Get remote URL
	if vcsBackend.RemoteURL == nil {
		return fmt.Errorf("migrate is not supported for this VCS backend")
	}
	remoteURL, err := vcsBackend.RemoteURL(absDir)
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Parse the remote URL
	u, err := newURL(remoteURL, false, false)
	if err != nil {
		return fmt.Errorf("failed to parse remote URL %q: %w", remoteURL, err)
	}

	// Derive destination path
	localRepo, err := LocalRepositoryFromURL(u, false)
	if err != nil {
		return fmt.Errorf("failed to derive destination path: %w", err)
	}

	destPath := localRepo.FullPath

	// Check if source and destination are the same
	if absDir == destPath {
		return fmt.Errorf("repository is already at the correct location: %s", destPath)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination directory %q already exists", destPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check destination directory: %w", err)
	}

	// Check for linked worktrees before dry-run return so we can report them
	var hasWorktrees bool
	if vcsBackend == GitBackend {
		hasWorktrees, err = hasLinkedWorktrees(absDir)
		if err != nil {
			return fmt.Errorf("failed to check for linked worktrees: %w", err)
		}
	}

	// Dry-run mode
	if dry {
		fmt.Fprintf(w, "Would migrate %s to %s\n", absDir, destPath)
		if hasWorktrees {
			fmt.Fprintf(w, "Would run 'git worktree repair' to update linked worktrees\n")
		}
		return nil
	}

	// Confirmation prompt (skip if -y flag is set)
	if !skipConfirm {
		ok, err := confirm(fmt.Sprintf("Migrate %s to %s?", absDir, destPath))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("migration aborted by user")
		}
	}

	// Create parent directories
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Move the repository
	if err := moveDir(absDir, destPath); err != nil {
		return fmt.Errorf("failed to move repository: %w", err)
	}

	// Repair linked worktrees so their .git files reference the new location.
	//
	// For each worktree, two pointers exist:
	//   back-pointer:    .git/worktrees/<name>/gitdir  → worktree working dir
	//   forward ref:     <worktree>/.git               → main repo's .git/worktrees/<name>
	//
	// External worktrees (outside the repo) didn't move, so only the forward
	// ref is stale. Internal worktrees (inside the repo) moved along with
	// the repo, so BOTH pointers are stale. We fix the back-pointers first
	// so that "git worktree repair" can match entries to update the forward refs.
	if hasWorktrees {
		wtPaths, wtErr := repairWorktreeBackPointers(absDir, destPath)
		if wtErr != nil {
			logger.Log("warning", fmt.Sprintf("failed to discover linked worktree paths: %v", wtErr))
		} else if len(wtPaths) > 0 {
			args := append([]string{"worktree", "repair"}, wtPaths...)
			cmd := exec.Command("git", args...)
			cmd.Dir = destPath
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Log("warning", fmt.Sprintf("git worktree repair failed: %v\n%s", err, out))
			}
		}
	}

	fmt.Fprintln(w, destPath)
	return nil
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
		if os.IsNotExist(err) {
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

	oldPrefix := oldDir + string(filepath.Separator)
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
		if strings.HasPrefix(wtPath, oldPrefix) {
			newPath := filepath.Join(destDir, wtPath[len(oldDir):])
			if err := os.WriteFile(gitdirFile, []byte(newPath+"\n"), 0644); err != nil {
				return nil, fmt.Errorf("failed to rewrite gitdir for worktree %s: %w", e.Name(), err)
			}
			wtPath = newPath
		}

		// The gitdir file stores the path to the worktree's .git file
		// (e.g., "/path/to/wt/.git"), but git worktree repair expects
		// the worktree working directory (e.g., "/path/to/wt").
		wtDir := strings.TrimSuffix(wtPath, string(filepath.Separator)+".git")
		paths = append(paths, wtDir)
	}
	return paths, nil
}

// moveDir attempts to move directory from src to dst, with fallback for cross-device moves
func moveDir(src, dst string) error {
	// Try atomic rename first
	renameErr := os.Rename(src, dst)
	if renameErr == nil {
		return nil
	}

	// Check for cross-device error
	var linkError *os.LinkError
	isCrossDevice := errors.As(renameErr, &linkError) && errors.Is(linkError.Err, syscall.EXDEV)

	if !isCrossDevice {
		return renameErr
	}

	// Fallback: copy directory tree using otiai10/copy, then remove source
	opt := copy.Options{
		// Preserve symlinks as-is
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Shallow
		},
		// Skip special files (pipes, sockets, devices) as they're uncommon in repos
		Skip: func(srcinfo os.FileInfo, src, dest string) (bool, error) {
			mode := srcinfo.Mode()
			// Skip if not regular file, directory, or symlink
			if !mode.IsRegular() && !mode.IsDir() && mode&os.ModeSymlink == 0 {
				return true, nil
			}
			return false, nil
		},
	}

	copyErr := copy.Copy(src, dst, opt)
	if copyErr != nil {
		// Attempt to cleanup partial copy
		if cleanupErr := os.RemoveAll(dst); cleanupErr != nil {
			return fmt.Errorf("copy failed: %w (cleanup also failed: %v)", copyErr, cleanupErr)
		}
		return copyErr
	}

	return os.RemoveAll(src)
}
