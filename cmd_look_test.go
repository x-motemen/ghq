package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/motemen/ghq/cmdutil"
)

func TestDoLook(t *testing.T) {
	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "ghq", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "gobump", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "Songmu", "gobump", ".git"), 0755)
		defer func(orig func(cmd *exec.Cmd) error) {
			cmdutil.CommandRunner = orig
		}(cmdutil.CommandRunner)
		var lastCmd *exec.Cmd
		cmdutil.CommandRunner = func(cmd *exec.Cmd) error {
			lastCmd = cmd
			return nil
		}
		sh := detectShell()

		err := newApp().Run([]string{"", "look", "https://github.com/motemen/ghq"})
		if err != nil {
			t.Errorf("error should be nil, but: %s", err)
		}

		if !reflect.DeepEqual(lastCmd.Args, []string{sh}) {
			t.Errorf("lastCmd.Args: got: %v, expect: %v", lastCmd.Args, []string{sh})
		}
		dir := filepath.Join(tmproot, "github.com", "motemen", "ghq")
		if filepath.Clean(lastCmd.Dir) != dir {
			t.Errorf("lastCmd.Dir: got: %s, expect: %s", lastCmd.Dir, dir)
		}
		gotEnv := lastCmd.Env[len(lastCmd.Env)-1]
		expectEnv := "GHQ_LOOK=github.com/motemen/ghq"
		if gotEnv != expectEnv {
			t.Errorf("lastCmd.Env[len(lastCmd.Env)-1]: got: %s, expect: %s", gotEnv, expectEnv)
		}

		err = newApp().Run([]string{"", "look", "github.com/motemen/_unknown"})
		expect := "No repository found"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}

		err = newApp().Run([]string{"", "look", "gobump"})
		expect = "More than one repositories are found; Try more precise name"
		if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
			t.Errorf("error should has prefix %q, but: %s", expect, err)
		}
	})
}
