package utils

import "testing"

func TestLog(t *testing.T) {
	Log("default", "shows this color")
	Log("error", "shows this color")
	Log("open", "shows this color")
	Log("authorized", "shows this color")
	Log("skip", "shows this color")
	Log("git", "shows this color")
}
