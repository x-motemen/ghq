package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

// GitConfigSingle fetches single git-config variable.
// returns an empty string and no error if no variable is found with the given key.
func GitConfigSingle(key string) (string, error) {
	return gitConfig("--get", key)
}

func GitConfigAll(key string) ([]string, error) {
	value, err := gitConfig("--get-all", key)
	if err != nil {
		return nil, err
	}

	var values = strings.Split(value, "\000")
	if len(values) == 1 && values[0] == "" {
		values = values[:0]
	}

	return values, nil
}

// GitConfigURLMatch emulates `git config --url-match <section>.<key>` with section containing dots.
// Used for "ghq.url.<URL>.<key>" config variables for example.
func GitConfigURLMatch(section, key, url string) (string, error) {
	keyRx := regexp.MustCompile(fmt.Sprintf(`^%s\.(.+)\.%s$`, strings.Replace(section, ".", `\.`, -1), key))
	pairs, err := gitConfig("--get-regexp", keyRx.String())
	if err != nil {
		return "", err
	}

	maxLength := -1
	value := ""
	for _, pair := range strings.Split(pairs, "\000") {
		// eg. {"ghq.url.https://ghe.example.com/.vcs", "github"}
		keyValue := strings.SplitN(pair, "\n", 2)
		if len(keyValue) < 2 {
			continue
		}

		m := keyRx.FindStringSubmatch(keyValue[0])
		if len(m) < 1 {
			continue
		}

		keyURL := m[1] // eg. "https://ghe.example.com/"

		if strings.HasPrefix(url, keyURL) {
			if len(keyURL) > maxLength {
				maxLength = len(keyURL)
				value = keyValue[1]
			}
		}
	}

	return value, nil
}

func gitConfig(getFlag, key string) (string, error) {
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
