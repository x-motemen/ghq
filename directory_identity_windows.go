//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func identifyDirectory(path string, _ os.FileInfo) (directoryIdentity, bool) {
	path, err := filepath.Abs(path)
	if err != nil {
		return directoryIdentity{}, false
	}
	if !strings.HasPrefix(path, `\\?\`) && !strings.HasPrefix(path, `\\.\`) {
		if strings.HasPrefix(path, `\\`) {
			path = `\\?\UNC\` + path[2:]
		} else {
			path = `\\?\` + path
		}
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return directoryIdentity{}, false
	}
	// No data access is needed, and aliases must be followed to their targets.
	handle, err := syscall.CreateFile(name, 0, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return directoryIdentity{}, false
	}
	defer syscall.CloseHandle(handle)
	var info syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &info); err != nil {
		return directoryIdentity{}, false
	}
	return directoryIdentity{uint64(info.VolumeSerialNumber), uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)}, true
}
