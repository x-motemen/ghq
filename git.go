package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func GitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--path", "--null", "--get", key)
	cmd.Stderr = os.Stderr

	buf, err := cmd.Output()

	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() == 1 {
				return "", nil
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return strings.TrimRight(string(buf), "\000"), nil
}
