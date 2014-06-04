package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func GitConfig(key string) (string, error) {
	return gitConfig(key, false)
}

func GitConfigAll(key string) ([]string, error) {
	value, err := gitConfig(key, true)
	if err != nil {
		return nil, err
	}

	var values = strings.Split(value, "\000")
	if len(values) == 1 && values[0] == "" {
		values = values[:0]
	}

	return values, nil
}

func gitConfig(key string, all bool) (string, error) {
	var getFlag string
	if all == true {
		getFlag = "--get-all"
	} else {
		getFlag = "--get"
	}

	cmd := exec.Command("git", "config", "--path", "--null", getFlag, key)
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
