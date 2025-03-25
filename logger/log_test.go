package logger

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	t.Run("with color", func(t *testing.T) {
		t.Logf("NO_COLOR: %s", os.Getenv("NO_COLOR"))
		selectLogger()
		// info
		Log("default", "should be green")
		// verbose
		Log("git", "should be white")
		Log("skip", "should be white")
		// notice
		Log("authorized", "should be blue")
		// warn
		Log("open", "should be yellow")
		// error
		Log("error", "should be red")
	})

	t.Run("without color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "true")
		t.Logf("NO_COLOR: %s", os.Getenv("NO_COLOR"))
		selectLogger()
		// info
		Log("default", "should be none")
		// verbose
		Log("git", "should be none")
		Log("skip", "should be none")
		// notice
		Log("authorized", "should be none")
		// warn
		Log("open", "should be none")
		// error
		Log("error", "should be none")
	})
}
