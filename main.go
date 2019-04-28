package main

import (
	"fmt"
	"os"

	"github.com/motemen/ghq/logger"
	"github.com/urfave/cli"
)

const Version = "0.10.2"

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
	app.Version = fmt.Sprintf("%s (rev:%s)", Version, revision)
	app.Author = "motemen"
	app.Email = "motemen@gmail.com"
	app.Commands = Commands
	return app
}
