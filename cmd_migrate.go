package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

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

	// Fallback: copy directory tree, then remove source
	copyErr := copyDir(src, dst)
	if copyErr != nil {
		// Attempt to cleanup partial copy, but prioritize returning the original error
		if cleanupErr := os.RemoveAll(dst); cleanupErr != nil {
			// Log cleanup failure but return the original copy error
			return fmt.Errorf("copy failed: %w (cleanup also failed: %v)", copyErr, cleanupErr)
		}
		return copyErr
	}

	return os.RemoveAll(src)
}

// copyDir recursively copies directory from src to dst, preserving permissions
func copyDir(src, dst string) error {
	walkFunc := func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			dirInfo, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			return os.MkdirAll(destPath, dirInfo.Mode().Perm())
		}

		// Get file info for type checking
		fileInfo, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}

		// Handle symbolic links
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			linkTarget, linkErr := os.Readlink(path)
			if linkErr != nil {
				return linkErr
			}
			return os.Symlink(linkTarget, destPath)
		}

		// Skip special non-regular files (named pipes, sockets, devices, etc.)
		if !fileInfo.Mode().IsRegular() {
			return nil
		}
		// Copy regular file
		return copyFile(path, destPath, fileInfo.Mode().Perm())
	}

	return filepath.WalkDir(src, walkFunc)
}

// copyFile copies a file from src to dst with specified permissions
func copyFile(src, dst string, perm os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
