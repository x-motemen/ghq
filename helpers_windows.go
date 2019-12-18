// +build windows

package main

import "syscall"

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
