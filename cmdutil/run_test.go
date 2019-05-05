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
