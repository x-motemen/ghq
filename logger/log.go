package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/motemen/go-colorine"
)

var (
	NoColor      = colorine.TextStyle{Foreground: colorine.None, Background: colorine.None}
	VerboseColor = colorine.Verbose // white
	InfoColor    = colorine.Info    // green
	NoticeColor  = colorine.Notice  // blue
	WarnColor    = colorine.Warn    // yellow
	ErrorColor   = colorine.Error   // red
)

var (
	logger = colorine.NewLogger( // default logger with color
		colorine.Prefixes{
			// verbose
			"git":      VerboseColor,
			"hg":       VerboseColor,
			"svn":      VerboseColor,
			"darcs":    VerboseColor,
			"pijul":    VerboseColor,
			"bzr":      VerboseColor,
			"fossil":   VerboseColor,
			"skip":     VerboseColor,
			"cd":       VerboseColor,
			"resolved": VerboseColor,
			// notice
			"authorized": NoticeColor,
			// warn
			"open":    WarnColor,
			"exists":  WarnColor,
			"warning": WarnColor,
			// error
			"error": ErrorColor,
		},
		InfoColor, // default is info
	)

	loggerWithoutColor = colorine.NewLogger(
		colorine.Prefixes{},
		NoColor,
	)
)

func init() {
	selectLogger()
}

func selectLogger() {
	if os.Getenv("NO_COLOR") != "" {
		logger = loggerWithoutColor
	}
	SetOutput(os.Stderr)
}

// SetOutput sets log output writer
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// Log outputs log
func Log(prefix, message string) {
	logger.Log(prefix, message)
}

// Logf outputs log with format
func Logf(prefix, msg string, args ...any) {
	Log(prefix, fmt.Sprintf(msg, args...))
}
