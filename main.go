package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/x-motemen/ghq/logger"
)

const version = "1.1.3"

var revision = "HEAD"

func main() {
	if err := newApp().Run(os.Args); err != nil {
		exitCode := 1
		if excoder, ok := err.(cli.ExitCoder); ok {
			exitCode = excoder.ExitCode()
		}
		logger.Log("error", err.Error())
		os.Exit(exitCode)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "ghq"
	app.Usage = "Manage remote repository clones"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, revision)
	app.Authors = []*cli.Author{{
		Name:  "motemen",
		Email: "motemen@gmail.com",
	}, {
		Name:  "Songmu",
		Email: "y.songmu@gmail.com",
	}}
	app.Commands = commands
	return app
}
