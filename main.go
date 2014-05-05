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
		{
			Name:   "get",
			Usage:  "Clone/sync with a remote repository",
			Action: CommandGet,
			Flags: []cli.Flag{
				cli.BoolFlag{"update, u", "Update local repository if cloned already"},
			},
		},
		{
			Name:   "list",
			Usage:  "List local repositories",
			Action: CommandList,
			Flags: []cli.Flag{
				cli.BoolFlag{"exact, e", "Perform an exact match"},
				cli.BoolFlag{"full-path, p", "Print full paths"},
				cli.BoolFlag{"unique", "Print unique subpaths"},
			},
		},
		{
			Name:   "look",
			Usage:  "Look into a local repository",
			Action: CommandLook,
		},
		{
			Name:   "pocket",
			Usage:  "Get for all github entries in Pocket",
			Action: CommandPocket,
			Flags: []cli.Flag{
				cli.BoolFlag{"update, u", "Update local repository if cloned already"},
			},
		},
	}

	app.Run(os.Args)
}
