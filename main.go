package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var Version string = "HEAD"

func main() {
	app := cli.NewApp()
	app.Name = "ghq"
	app.Usage = "Manage GitHub repository clones"
	app.Version = Version
	app.Author = "motemen"
	app.Email = "motemen@gmail.com"
	app.Commands = Commands
	app.Run(os.Args)
}
