package main

import (
	"fmt"
	"os/exec"
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
