package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
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
