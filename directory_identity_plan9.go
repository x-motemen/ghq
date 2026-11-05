//go:build plan9

package main

import (
	"os"
	"syscall"
)

func identifyDirectory(_ string, info os.FileInfo) (directoryIdentity, bool) {
	stat, ok := info.Sys().(*syscall.Dir)
	if !ok {
		return directoryIdentity{}, false
	}
	return directoryIdentity{uint64(stat.Type)<<32 | uint64(stat.Dev), stat.Qid.Path}, true
}
