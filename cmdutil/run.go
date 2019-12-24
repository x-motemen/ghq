package cmdutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/motemen/ghq/logger"
)

// Run the command
func Run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return RunCommand(cmd, false)
}

// RunSilently runs the command silently
func RunSilently(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	return RunCommand(cmd, true)
}

// RunInDir runs the command in the specified directory
func RunInDir(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	return RunCommand(cmd, false)
}

// RunInDirSilently run the command in the specified directory silently
func RunInDirSilently(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	cmd.Dir = dir

	return RunCommand(cmd, true)
}

// RunInDirStderr run the command in the specified directory and prevent stdout output
func RunInDirStderr(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	return RunCommand(cmd, true)
}

// RunFunc for the type command execution
type RunFunc func(*exec.Cmd) error

// CommandRunner is for running the command
var CommandRunner = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// RunCommand run the command
func RunCommand(cmd *exec.Cmd, silent bool) error {
	if !silent {
		logger.Log(cmd.Args[0], strings.Join(cmd.Args[1:], " "))
	}
	err := CommandRunner(cmd)
	if err != nil {
		return &RunError{cmd, err}
	}

	return nil
}

// RunError is the error type for cmdutil
type RunError struct {
	Command   *exec.Cmd
	ExecError error
}

// Error to implement error interface
func (e *RunError) Error() string {
	return fmt.Sprintf("%s: %s", e.Command.Path, e.ExecError)
}
