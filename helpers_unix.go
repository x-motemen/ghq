//go:build !windows

package main

import "path/filepath"

func toFullPath(s string) (string, error) {
	return s, nil
}

func evalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
