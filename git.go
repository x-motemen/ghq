package main

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/motemen/ghq/logger"
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

var versionRx = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

var featureConfigURLMatchVersion = []uint{1, 8, 5}

func GitHasFeatureConfigURLMatch() bool {
	cmd := exec.Command("git", "--version")
	buf, err := cmd.Output()

	if err != nil {
		return false
	}

	return gitVersionOutputSatisfies(string(buf), featureConfigURLMatchVersion)
}

func gitVersionOutputSatisfies(gitVersionOutput string, baseVersionParts []uint) bool {
	versionStrings := versionRx.FindStringSubmatch(gitVersionOutput)
	if versionStrings == nil {
		return false
	}

	for i, v := range baseVersionParts {
		thisV64, err := strconv.ParseUint(versionStrings[i+1], 10, 0)
		logger.PanicIf(err)

		thisV := uint(thisV64)

		if thisV > v {
			return true
		} else if v == thisV {
			continue
		} else {
			return false
		}
	}

	return true
}
