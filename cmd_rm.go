package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/x-motemen/ghq/logger"
)

func doRm(ctx context.Context, cmd *cli.Command) error {
	var (
		name = cmd.Args().First()
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

	localRepo, err := LocalRepositoryFromURL(u, bare)
	if err != nil {
		return err
	}

	p := localRepo.FullPath
	ok, err := isNotExistOrEmpty(p)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("directory %q does not exist", p)
	}

	// Scenario A: Is this path itself a linked worktree?
	isWorktree := false
	var gitdirTarget string
	if linked, target, linkErr := isLinkedGitDir(p); linkErr != nil {
		return fmt.Errorf("failed to check worktree status: %w", linkErr)
	} else if linked && isWorktreeGitDir(target) {
		isWorktree = true
		gitdirTarget = target
	}

	// Scenario B: Does this repo have linked worktrees?
	var worktreePaths []string
	if !isWorktree {
		if hasWt, wtErr := hasLinkedWorktrees(p); wtErr != nil {
			return fmt.Errorf("failed to check for linked worktrees: %w", wtErr)
		} else if hasWt {
			worktreePaths, err = listLinkedWorktreePaths(p)
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

	ok, err = confirm(confirmMsg)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("aborted")
	}

	// Removal
	if isWorktree {
		// Use git worktree remove to properly unregister from parent repo.
		// Run from inside the worktree so git can read its .git file to
		// discover the main repository.
		gitCmd := exec.Command("git", "worktree", "remove", "--force", p)
		gitCmd.Dir = p
		if out, gitErr := gitCmd.CombinedOutput(); gitErr != nil {
			logger.Log("warning", fmt.Sprintf("git worktree remove failed: %v\n%s", gitErr, out))
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
			gitCmd := exec.Command("git", "worktree", "remove", "--force", wt)
			gitCmd.Dir = p
			if out, gitErr := gitCmd.CombinedOutput(); gitErr != nil {
				logger.Log("warning", fmt.Sprintf("failed to remove worktree %s: %v\n%s", wt, gitErr, out))
			}
		}
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	fmt.Fprintf(w, "Removed %s\n", p)
	return nil
}

func confirm(message string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, err
	}
	return response == "y", nil
}
