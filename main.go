package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/x-motemen/ghq/logger"
)

const version = "1.10.2"

var revision = "HEAD"

func main() {
	if err := newApp().Run(context.Background(), os.Args); err != nil {
		exitCode := 1
		if excoder, ok := err.(cli.ExitCoder); ok {
			exitCode = excoder.ExitCode()
		}
		logger.Log("error", err.Error())
		os.Exit(exitCode)
	}
}

func newApp() *cli.Command {
	return &cli.Command{
		Name:     "ghq",
		Usage:    "Manage remote repository clones",
		Version:  fmt.Sprintf("%s (rev:%s)", version, revision),
		Authors:  []any{"motemen <motemen@gmail.com>", "Songmu <y.songmu@gmail.com>"},
		Suggest:  true,
		Commands: commands,
	}
}
