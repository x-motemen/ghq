//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func toFullPath(s string) (string, error) {
	p := syscall.StringToUTF16(s)
	b := p
	n, err := syscall.GetLongPathName(&p[0], &b[0], uint32(len(b)))
	if err != nil {
		return s, err
	}
	if n > uint32(len(b)) {
		b = make([]uint16, n)
		n, err = syscall.GetLongPathName(&p[0], &b[0], uint32(len(b)))
		if err != nil {
			return s, err
		}
	}
	b = b[:n]
	return syscall.UTF16ToString(b), nil
}

func evalSymlinks(path string) (string, error) {
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	list := filepathSplitAll(path)
	evaled := list[0]
	for i := 1; i < len(list); i++ {
		evaled = filepath.Join(evaled, list[i])

		linkSrc, err := os.Readlink(evaled)
		if err != nil {
			// not symlink
			continue
		} else {
			if filepath.IsAbs(linkSrc) {
				evaled = linkSrc
			} else {
				evaled = filepath.Join(filepath.Dir(evaled), linkSrc)
			}
		}
	}

	return evaled, nil
}

func filepathSplitAll(path string) []string {
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)

	vol := filepath.VolumeName(path)

	path = path[len(vol):]
	list := strings.Split(path, "/")
	list[0] = vol + string(filepath.Separator) + list[0]
	return list
}
