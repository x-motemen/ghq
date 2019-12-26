// +build !windows

package main

import (
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func stdinIsPipe() bool {
	return !terminal.IsTerminal(syscall.Stdin)
}
