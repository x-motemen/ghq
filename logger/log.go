package logger

import (
	"os"

	"github.com/motemen/go-colorine"
)

var logger = colorine.NewLogger(
	colorine.Prefixes{
		"git":      colorine.Verbose,
		"hg":       colorine.Verbose,
		"svn":      colorine.Verbose,
		"darcs":    colorine.Verbose,
		"skip":     colorine.Verbose,
		"cd":       colorine.Verbose,
		"resolved": colorine.Verbose,

		"open":    colorine.Warn,
		"exists":  colorine.Warn,
		"warning": colorine.Warn,

		"authorized": colorine.Notice,

		"error": colorine.Error,
	}, colorine.Info)

func init() {
	logger.SetOutput(os.Stderr)
}

func Log(prefix, message string) {
	logger.Log(prefix, message)
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}
