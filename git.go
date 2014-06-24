package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// GitConfigSingle fetches single git-config variable.
// returns an empty string and no error if no variable is found with the given key.
func GitConfigSingle(key string) (string, error) {
	return GitConfig("--get", key)
}

// GitConfigAll fetches git-config variable of multiple values.
func GitConfigAll(key string) ([]string, error) {
	value, err := GitConfig("--get-all", key)
	if err != nil {
		return nil, err
	}

	// No results found, return an empty slice
	if value == "" {
		return nil, nil
	}

	return strings.Split(value, "\000"), nil
}

// GitConfig invokes 'git config' and handles some errors properly.
func GitConfig(args ...string) (string, error) {
	gitArgs := append([]string{"config", "--path", "--null"}, args...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Stderr = os.Stderr

	buf, err := cmd.Output()

	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() == 1 {
				// The key was not found, do not treat as an error
				return "", nil
			}
		}

		return "", err
	}

	return strings.TrimRight(string(buf), "\000"), nil
}
