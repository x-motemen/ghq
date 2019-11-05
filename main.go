package main

import (
	"fmt"
	"os"

	"github.com/motemen/ghq/logger"
	"github.com/urfave/cli"
)

const version = "0.12.8"

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
	app.Usage = "Manage GitHub repository clones"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, revision)
	app.Author = "motemen"
	app.Email = "motemen@gmail.com"
	app.Commands = commands
	return app
}
