//go:build !windows

package main

func toFullPath(s string) (string, error) {
	return s, nil
}
