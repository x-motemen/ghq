package main

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/motemen/ghq/utils"
)

// Represents a VCS backend.
type VCSBackend struct {
	// Clones a remote repository to local path.
	Clone func(*url.URL, string) error
	// Updates a cloned local repository.
	Update func(string) error
}

var GitBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return utils.Run("git", "clone", remote.String(), local)
	},
	Update: func(local string) error {
		err := os.Chdir(local)
		if err != nil {
			return err
		}

		return utils.Run("git", "remote", "update")
	},
}

var MercurialBackend = &VCSBackend{
	Clone: func(remote *url.URL, local string) error {
		dir, _ := filepath.Split(local)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		return utils.Run("hg", "clone", remote.String(), local)
	},
	Update: func(local string) error {
		err := os.Chdir(local)
		if err != nil {
			return err
		}

		return utils.Run("hg", "pull")
	},
}
