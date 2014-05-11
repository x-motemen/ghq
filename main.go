package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ghq"
	app.Usage = "Manage GitHub repository clones"
	app.Version = "0.1.0"
	app.Author = "motemen"
	app.Email = "motemen@gmail.com"
	app.Commands = []cli.Command{
		GetCommand,
		ListCommand,
		LookCommand,
		PocketCommand,
	}

	app.Run(os.Args)
}
