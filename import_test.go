package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/motemen/ghq/logger"
)

func TestDoImport(t *testing.T) {
	in := []string{
		"github.com/motemen/ghq",
		"github.com/motemen/gore",
	}

	testCases := []struct {
		name string
		args []string
	}{{
		name: "normal",
		args: []string{},
	}, {
		name: "parallel (Experimental)",
		args: []string{"-P"},
	}}

	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	defer func() { logger.SetOutput(os.Stderr) }()

	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		for _, r := range in {
			os.MkdirAll(filepath.Join(tmproot, r, ".git"), 0755)
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf.Reset()
				out, _, err := captureWithInput(in, func() {
					args := append([]string{"", "import"}, tc.args...)
					if err := newApp().Run(args); err != nil {
						t.Errorf("error should be nil but: %s", err)
					}
				})
				if err != nil {
					t.Errorf("error should be nil, but: %s", err)
				}
				if out != "" {
					t.Errorf("out should be empty, but: %s", out)
				}
				log := filepath.ToSlash(buf.String())
				for _, r := range in {
					if !strings.Contains(log, r) {
						t.Errorf("log should contains %q but not: %s", r, log)
					}
				}
			})
		}
	})
}
