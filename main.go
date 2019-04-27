package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const Version = "0.9.0"

var revision = "HEAD"

func main() {
	newApp().Run(os.Args)
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
