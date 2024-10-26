//go:build windows

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type testEvalSymlinksMode int

const (
	testEvalSymlinksNotLink testEvalSymlinksMode = iota
	testEvalSymlinksSymbolicLink
	testEvalSymlinksJunction
)

func Test_evalSymlinks(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name         string
		mode         testEvalSymlinksMode
		linkBasePath string
		args         args
		want         string
		wantErr      bool
	}{
		{
			name: "not link",
			mode: testEvalSymlinksNotLink,
			args: args{
				path: filepath.Join(os.TempDir(), "not_link"),
			},
			want:    filepath.Join(os.TempDir(), "not_link"),
			wantErr: false,
		},
		{
			name:         "symbolic link",
			mode:         testEvalSymlinksSymbolicLink,
			linkBasePath: filepath.Join(os.TempDir(), "link_base"),
			args: args{
				path: filepath.Join(os.TempDir(), "symbolic_link"),
			},
			want:    filepath.Join(os.TempDir(), "link_base"),
			wantErr: false,
		},
		{
			name:         "junction",
			mode:         testEvalSymlinksJunction,
			linkBasePath: filepath.Join(os.TempDir(), "link_base"),
			args: args{
				path: filepath.Join(os.TempDir(), "junction"),
			},
			want:    filepath.Join(os.TempDir(), "link_base"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createLink(tt.linkBasePath, tt.args.path, tt.mode); err != nil {
				t.Errorf("failed to create link: %v", err)
				return
			}

			got, err := evalSymlinks(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalSymlinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evalSymlinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createLink(linkBasePath, path string, mode testEvalSymlinksMode) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	if mode == testEvalSymlinksNotLink {
		return os.MkdirAll(path, 0755)
	}

	if err := os.MkdirAll(linkBasePath, 0755); err != nil {
		return err
	}

	switch mode {
	case testEvalSymlinksSymbolicLink:
		return os.Symlink(linkBasePath, path)
	case testEvalSymlinksJunction:
		output, err := exec.Command("cmd", "/c", "mklink", "/J", path, linkBasePath).CombinedOutput()
		if err != nil {
			output, err := io.ReadAll(transform.NewReader(bytes.NewBuffer(output), japanese.ShiftJIS.NewDecoder()))
			if err != nil {
				return fmt.Errorf("failed to transform output: %w", err)
			}
			return fmt.Errorf("failed to create junction: %s, %w", string(output), err)
		}
		return nil
	}
	return nil
}
