package cmdutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestRunInDirSilently(t *testing.T) {
	err := RunInDirSilently(".", "/path/to/unknown")
	expect := "/path/to/unknown: "
	if !strings.HasPrefix(fmt.Sprintf("%s", err), expect) {
		t.Errorf("error message should have prefix %q, but: %q", expect, err)
	}
}

func TestRun(t *testing.T) {
	err := Run("echo")
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
}

func TestRunInDir(t *testing.T) {
	err := RunInDir(".", "echo")
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
}

func TestRunSilently(t *testing.T) {
	err := RunSilently("echo")
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
}
