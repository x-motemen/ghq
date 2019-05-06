package gitutil

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/blang/semver"
)

// ConfigSingle fetches single git-config variable.
// returns an empty string and no error if no variable is found with the given key.
func ConfigSingle(key string) (string, error) {
	return Config("--get", key)
}

// ConfigAll fetches git-config variable of multiple values.
func ConfigAll(key string) ([]string, error) {
	value, err := Config("--get-all", key)
	if err != nil {
		return nil, err
	}

	// No results found, return an empty slice
	if value == "" {
		return nil, nil
	}

	return strings.Split(value, "\000"), nil
}

// Config invokes 'git config' and handles some errors properly.
func Config(args ...string) (string, error) {
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

var (
	versionRx                    = regexp.MustCompile(`((?:\d+)\.(?:\d+)\.(?:\d+))`)
	featureConfigURLMatchVersion = semver.MustParse("1.8.5")
)

// HasFeatureConfigURLMatch checks has url-match feature or not
func HasFeatureConfigURLMatch() error {
	cmd := exec.Command("git", "--version")
	buf, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("failed to execute %q: %s", "git --version", err)
	}

	return gitVersionOutputSatisfies(string(buf), featureConfigURLMatchVersion)
}

func gitVersionOutputSatisfies(gitVersionOutput string, baseVersion semver.Version) error {
	versionStrings := versionRx.FindStringSubmatch(gitVersionOutput)
	if len(versionStrings) == 0 {
		return fmt.Errorf("failed to detect git version from %q", gitVersionOutput)
	}
	ver, err := semver.Parse(versionStrings[1])
	if err != nil {
		return fmt.Errorf("failed to parse version string %q: %s", versionStrings[1], err)
	}
	if ver.LT(baseVersion) {
		return fmt.Errorf("This version of Git does not support `config --get-urlmatch`; per-URL settings are not available")
	}
	return nil
}
