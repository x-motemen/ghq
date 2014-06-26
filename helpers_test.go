package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/motemen/ghq/utils"
)

func NewFakeRunner(dispatch map[string]error) utils.RunFunc {
	return func(cmd *exec.Cmd) error {
		cmdString := strings.Join(cmd.Args, " ")
		for cmdPrefix, err := range dispatch {
			if strings.Index(cmdString, cmdPrefix) == 0 {
				return err
			}
		}
		panic(fmt.Sprintf("No fake dispatch found for: %s", cmdString))
	}
}

func WithGitconfigFile(configContent string) (func(), error) {
	tmpdir, err := ioutil.TempDir("", "ghq-test")
	if err != nil {
		return nil, err
	}

	tmpGitconfigFile := filepath.Join(tmpdir, "gitconfig")

	ioutil.WriteFile(
		tmpGitconfigFile,
		[]byte(configContent),
		0777,
	)

	prevGitConfigEnv := os.Getenv("GIT_CONFIG")
	os.Setenv("GIT_CONFIG", tmpGitconfigFile)

	return func() {
		os.Setenv("GIT_CONFIG", prevGitConfigEnv)
	}, nil
}
