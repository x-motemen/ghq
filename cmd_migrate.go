package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/otiai10/copy"
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
	if err := moveDir(absDir, destPath); err != nil {
		return fmt.Errorf("failed to move repository: %w", err)
	}

	fmt.Fprintln(w, destPath)
	return nil
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
