package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func doMigrate(c *cli.Context) error {
	var (
		repoDir = c.Args().First()
		dry     = c.Bool("dry-run")
		skipConfirm = c.Bool("y")
		w       = c.App.Writer
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

	// Get remote URL (currently only git is supported)
	var remoteURL string
	if vcsBackend == GitBackend {
		remoteURL, err = getGitRemoteURL(absDir)
		if err != nil {
			return fmt.Errorf("failed to get remote URL: %w", err)
		}
	} else {
		return fmt.Errorf("only git repositories are currently supported")
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

	// Dry-run mode
	if dry {
		fmt.Fprintf(w, "Would migrate %s to %s\n", absDir, destPath)
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
	if err := os.Rename(absDir, destPath); err != nil {
		return fmt.Errorf("failed to move repository: %w", err)
	}

	fmt.Fprintln(w, destPath)
	return nil
}

// getGitRemoteURL retrieves the remote URL from a git repository
// It tries 'origin' first, then falls back to the first remote
func getGitRemoteURL(repoDir string) (string, error) {
	// Try to get the URL of 'origin' remote
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(output))
		if url != "" {
			return url, nil
		}
	}

	// Fall back to the first remote
	cmd = exec.Command("git", "remote")
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list remotes: %w", err)
	}

	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(remotes) == 0 || remotes[0] == "" {
		return "", fmt.Errorf("no remotes found")
	}

	// Get URL of the first remote
	firstRemote := remotes[0]
	cmd = exec.Command("git", "remote", "get-url", firstRemote)
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get URL of remote %q: %w", firstRemote, err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("remote %q has no URL", firstRemote)
	}

	return url, nil
}
