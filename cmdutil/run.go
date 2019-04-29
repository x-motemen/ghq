package cmdutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/motemen/ghq/logger"
)

func Run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return RunCommand(cmd)
}

func RunSilently(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	return RunCommand(cmd)
}

func RunInDir(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	return RunCommand(cmd)
}

func RunInDirSilently(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	cmd.Dir = dir

	return RunCommand(cmd)
}

type RunFunc func(*exec.Cmd) error

var CommandRunner = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

func RunCommand(cmd *exec.Cmd) error {
	logger.Log(cmd.Args[0], strings.Join(cmd.Args[1:], " "))

	err := CommandRunner(cmd)
	if err != nil {
		if execErr, ok := err.(*exec.Error); ok {
			logger.Log("warning", fmt.Sprintf("%q: %s", execErr.Name, execErr.Err))
		}
		return &RunError{cmd, err}
	}

	return nil
}

type RunError struct {
	Command   *exec.Cmd
	ExecError error
}

func (e *RunError) Error() string {
	return fmt.Sprintf("%s: %s", e.Command.Path, e.ExecError)
}
