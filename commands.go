package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"
)

var commands = []*cli.Command{
	commandGet,
	commandList,
	commandRm,
	commandRoot,
	commandCreate,
	commandCheck,
}

var commandGet = &cli.Command{
	Name:	"get",
	Aliases: []string{"clone"},
	Usage:	"Clone/sync with a remote repository",
	Description: `
    Clone a repository under ghq root directory. If the repository is
    already cloned to local, nothing will happen unless '-u' ('--update')
    flag is supplied, in which case 'git remote update' is executed.
    When you use '-p' option, the repository is cloned via SSH.`,
	Action: doGet,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "update", Aliases: []string{"u"},
			Usage: "Update local repository if cloned already"},
		&cli.BoolFlag{Name: "p", Usage: "Clone with SSH"},
		&cli.BoolFlag{Name: "shallow", Usage: "Do a shallow clone"},
		&cli.BoolFlag{Name: "look", Aliases: []string{"l"}, Usage: "Look after get"},
		&cli.StringFlag{Name: "vcs", Usage: "Specify `vcs` backend for cloning"},
		&cli.BoolFlag{Name: "silent", Aliases: []string{"s"}, Usage: "clone or update silently"},
		&cli.BoolFlag{Name: "no-recursive", Usage: "prevent recursive fetching"},
		&cli.StringFlag{Name: "branch", Aliases: []string{"b"},
			Usage: "Specify `branch` name. This flag implies --single-branch on Git"},
		&cli.BoolFlag{Name: "parallel", Aliases: []string{"P"}, Usage: "Import parallelly"},
		&cli.BoolFlag{Name: "bare", Usage: "Do a bare clone"},
		&cli.StringFlag{
			Name:	"partial",
			Usage: "Do a partial clone. Can specify either \"blobless\" or \"treeless\"",
			Action: func(ctx *cli.Context, v string) error {
				expected := []string{"blobless", "treeless"}
				if !slices.Contains(expected, v) {
					return fmt.Errorf("flag partial value \"%v\" is not allowed", v)
				}
				return nil
			}},
	},
}

var commandList = &cli.Command{
	Name:	"list",
	Usage: "List local repositories",
	Description: `
    List locally cloned repositories. If a query argument is given, only
    repositories whose names contain that query text are listed.
    '-e' ('--exact') forces the match to be an exact one (i.e. the query equals to
    project or user/project) If '-p' ('--full-path') is given, the full paths
    to the repository root are printed instead of relative ones.`,
	Action: doList,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "exact", Aliases: []string{"e"}, Usage: "Perform an exact match"},
		&cli.StringFlag{Name: "vcs", Usage: "Specify `vcs` backend for matching"},
		&cli.BoolFlag{Name: "full-path", Aliases: []string{"p"}, Usage: "Print full paths"},
		&cli.BoolFlag{Name: "unique", Usage: "Print unique subpaths"},
		&cli.BoolFlag{Name: "bare", Usage: "Query bare repositories"},
	},
}

var commandCheck = &cli.Command{
	Name:	"check",
	Usage:	"Check for uncommitted changes, stashes, and untracked files",
	Action: doCheck,
}

var commandRm = &cli.Command{
	Name:	"rm",
	Usage:	"Remove local repository",
	Action: doRm,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "dry-run", Usage: "Do not remove actually"},
	},
}

var commandRoot = &cli.Command{
	Name:	"root",
	Usage:	"Show repositories' root",
	Action: doRoot,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "all", Usage: "Show all roots"},
	},
}

var commandCreate = &cli.Command{
	Name:	"create",
	Usage:	"Create a new repository",
	Action: doCreate,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "vcs", Usage: "Specify `vcs` backend explicitly"},
		&cli.BoolFlag{Name: "bare", Usage: "Create a bare repository"},
	},
}

type commandDoc struct {
	Parent		string
	Arguments	string
}

var commandDocs = map[string]commandDoc{
	"get":	{"", "[-u] [-p] [--shallow] [--vcs <vcs>] [--look] [--silent] [--branch <branch>] [--no-recursive] [--bare] [--partial blobless|treeless] <repository URL>|<project>|<user>/<project>|<host>/<user>/<project>"},
	"list":	{"", "[-p] [-e] [<query>]"},
	"check": {"", ""},
	"create": {"", "<project>|<user>/<project>|<host>/<user>/<project>"},
	"rm":	{"", "<project>|<user>/<project>|<host>/<user>/<project>"},
	"root":	{"", "[-all]"},
}

// Makes template conditionals to generate per-command documents.
func mkCommandsTemplate(genTemplate func(commandDoc) string) string {
	template := "{{if false}}"
	for _, command := range commands {
		template = template + fmt.Sprintf("{{else if (eq .Name %q)}}%s", command.Name, genTemplate(commandDocs[command.Name]))
	}
	return template + "{{end}}"
}

func init() {
	argsTemplate := mkCommandsTemplate(func(doc commandDoc) string { return doc.Arguments })
	parentTemplate := mkCommandsTemplate(func(doc commandDoc) string { return strings.TrimLeft(doc.Parent+" ", " ") })

	cli.CommandHelpTemplate = `NAME:
    {{.Name}} - {{.Usage}}

USAGE:
    ghq ` + parentTemplate + `{{.Name}} ` + argsTemplate + `
{{if (len .Description)}}
DESCRIPTION: {{.Description}}
{{end}}{{if (len .Flags)}}
OPTIONS:
    {{range .Flags}}{{.}}
    {{end}}
{{end}}`
}
