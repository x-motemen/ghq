package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/x-motemen/ghq/logger"
)

func doRm(ctx context.Context, cmd *cli.Command) error {
	names := cmd.Args().Slice()
	if len(names) == 0 {
		return fmt.Errorf("repository name is required")
	}
	for _, name := range names {
		if err := removeRepository(cmd, name); err != nil {
			return err
		}
	}
	return nil
}

func removeRepository(cmd *cli.Command, name string) error {
	var (
		dry  = cmd.Bool("dry-run")
		w    = cmd.Root().Writer
		bare = cmd.Bool("bare")
	)

	if name == "" {
		return fmt.Errorf("repository name is required")
	}

	u, err := newURL(name, false, true)
	if err != nil {
		return err
	}
	// Apply host-specific rules. Unknown hosts may use one path component.
	if _, err := NewRemoteRepository(u); err != nil {
		return err
	}

	p, isAlias, err := resolveRemovalEntry(u, bareModeFromClassicBool(bare))
	if err != nil {
		return err
	}

	// Scenario A: Is this path itself a linked worktree?
	isWorktree := false
	var gitdirTarget string
	if !isAlias {
		if linked, target, linkErr := isLinkedGitDir(p); linkErr != nil {
			return fmt.Errorf("failed to check worktree status: %w", linkErr)
		} else if linked && isWorktreeGitDir(target) {
			isWorktree = true
			gitdirTarget = target
		}
	}

	// Scenario B: Does this repo have linked worktrees?
	var worktreePaths []string
	gitDir := filepath.Join(p, ".git")
	if !isWorktree && !isAlias {
		if bare {
			// --bare selects a path suffix, not the repository's Git layout.
			if _, statErr := os.Stat(gitDir); os.IsNotExist(statErr) {
				gitDir = p
			} else if statErr != nil {
				return fmt.Errorf("failed to check git directory: %w", statErr)
			}
		}
		if hasWt, wtErr := hasLinkedWorktrees(gitDir); wtErr != nil {
			return fmt.Errorf("failed to check for linked worktrees: %w", wtErr)
		} else if hasWt {
			worktreePaths, err = listLinkedWorktreePaths(gitDir)
			if err != nil {
				return fmt.Errorf("failed to list linked worktrees: %w", err)
			}
		}
	}

	// Dry-run
	if dry {
		if isWorktree {
			fmt.Fprintf(w, "Would remove worktree %s (linked to %s)\n", p, gitdirTarget)
		} else if len(worktreePaths) > 0 {
			fmt.Fprintf(w, "Would remove %s and its %d linked worktree(s):\n", p, len(worktreePaths))
			for _, wt := range worktreePaths {
				fmt.Fprintf(w, "  %s\n", wt)
			}
		} else {
			fmt.Fprintf(w, "Would remove %s\n", p)
		}
		return nil
	}

	// Confirmation
	var confirmMsg string
	if isWorktree {
		confirmMsg = fmt.Sprintf("Remove worktree %s?", p)
	} else if len(worktreePaths) > 0 {
		confirmMsg = fmt.Sprintf("Remove %s and its %d linked worktree(s)?\n  %s",
			p, len(worktreePaths), strings.Join(worktreePaths, "\n  "))
	} else {
		confirmMsg = fmt.Sprintf("Remove %s?", p)
	}

	ok, err := confirm(confirmMsg)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("aborted")
	}

	// Removal
	if isAlias {
		if err := os.Remove(p); err != nil {
			return err
		}
	} else if isWorktree {
		// Use git worktree remove to properly unregister from parent repo.
		// Resolve the main repo directory so we don't run git from inside
		// the directory being deleted.
		removed := false
		if mainRepoDir, dirErr := resolveMainRepoDir(gitdirTarget); dirErr == nil {
			gitCmd := exec.Command("git", "worktree", "remove", "--force", p)
			gitCmd.Dir = mainRepoDir
			if out, gitErr := gitCmd.CombinedOutput(); gitErr != nil {
				logger.Log("warning", fmt.Sprintf("git worktree remove failed: %v\n%s", gitErr, out))
			} else {
				removed = true
			}
		} else {
			logger.Log("warning", fmt.Sprintf("cannot resolve main repo dir: %v", dirErr))
		}
		if !removed {
			logger.Log("warning", "falling back to direct removal")
			if err := os.RemoveAll(p); err != nil {
				return err
			}
			// Best-effort cleanup of dangling .git/worktrees/<name> entry
			if gitdirTarget != "" {
				os.RemoveAll(gitdirTarget)
			}
		}
	} else {
		// Prune linked worktrees before removing main repo
		for _, wt := range worktreePaths {
			if _, statErr := os.Stat(wt); os.IsNotExist(statErr) {
				continue // already gone
			}
			gitCmd := exec.Command("git", "--git-dir", gitDir, "worktree", "remove", "--force", wt)
			gitCmd.Dir = p
			if out, gitErr := gitCmd.CombinedOutput(); gitErr != nil {
				return fmt.Errorf("failed to remove worktree %s: %w\n%s", wt, gitErr, out)
			}
		}
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	fmt.Fprintf(w, "Removed %s\n", p)
	return nil
}

func resolveRemovalEntry(u *url.URL, mode BareMode) (string, bool, error) {
	repo, found, err := resolveLocalRepository(u, mode)
	if err != nil {
		return "", false, err
	}
	p := repo.FullPath
	fi, err := os.Lstat(p)
	isAlias := err == nil && isLinkEntry(p, fi)
	if !found && !isAlias {
		roots, rootsErr := localRepositoryRoots(true)
		if rootsErr != nil {
			return "", false, rootsErr
		}
		for _, root := range roots {
			candidate := filepath.Join(root, repo.RelPath)
			entry, statErr := os.Lstat(candidate)
			if os.IsNotExist(statErr) || os.IsPermission(statErr) || isNotADirectory(statErr) {
				continue
			}
			if statErr != nil {
				return "", false, statErr
			}
			if isLinkEntry(candidate, entry) {
				return candidate, true, nil
			}
		}
	}
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, fmt.Errorf("directory %q does not exist", p)
		}
		return "", false, err
	}
	if isAlias {
		return p, true, nil
	}
	empty, err := isNotExistOrEmpty(p)
	if err != nil {
		return "", false, err
	}
	if empty {
		return "", false, fmt.Errorf("directory %q does not exist", p)
	}
	return p, false, nil
}

func confirm(message string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, err
	}
	return response == "y", nil
}
