package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/motemen/ghq/utils"
)

func Git(command ...string) {
	utils.Log("git", strings.Join(command, " "))

	cmd := exec.Command("git", command...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		utils.Log("error", fmt.Sprintf("git: %s", err))
		os.Exit(1)
	}
}

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
