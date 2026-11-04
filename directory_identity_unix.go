//go:build !windows && !plan9

package main

import (
	"os"
	"syscall"
)

func identifyDirectory(_ string, info os.FileInfo) (directoryIdentity, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return directoryIdentity{}, false
	}
	return directoryIdentity{uint64(stat.Dev), uint64(stat.Ino)}, true
}
