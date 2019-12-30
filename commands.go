package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

var commands = []*cli.Command{
	commandGet,
	commandList,
	commandRoot,
	commandCreate,
}

var commandGet = &cli.Command{
	Name:  "get",
	Usage: "Clone/sync with a remote repository",
	Description: `
    Clone a GitHub repository under ghq root directory. If the repository is
    already cloned to local, nothing will happen unless '-u' ('--update')
    flag is supplied, in which case 'git remote update' is executed.
    When you use '-p' option, the repository is cloned via SSH.
`,
	Action: doGet,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "update", Aliases: []string{"u"},
			Usage: "Update local repository if cloned already"},
		&cli.BoolFlag{Name: "p", Usage: "Clone with SSH"},
		&cli.BoolFlag{Name: "shallow", Usage: "Do a shallow clone"},
		&cli.BoolFlag{Name: "look", Aliases: []string{"l"}, Usage: "Look after get"},
		&cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for cloning"},
		&cli.BoolFlag{Name: "silent", Aliases: []string{"s"}, Usage: "clone or update silently"},
		&cli.BoolFlag{Name: "no-recursive", Usage: "prevent recursive fetching"},
		&cli.StringFlag{Name: "branch", Aliases: []string{"b"},
			Usage: "Specify branch name. This flag implies --single-branch on Git"},
		&cli.BoolFlag{Name: "parallel", Aliases: []string{"P"}, Usage: "Import parallely"},
	},
}

var commandList = &cli.Command{
	Name:  "list",
	Usage: "List local repositories",
	Description: `
    List locally cloned repositories. If a query argument is given, only
    repositories whose names contain that query text are listed. '-e'
    ('--exact') forces the match to be an exact one (i.e. the query equals to
    _project_ or _user_/_project_) If '-p' ('--full-path') is given, the full paths
    to the repository root are printed instead of relative ones.
`,
	Action: doList,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "exact", Aliases: []string{"e"}, Usage: "Perform an exact match"},
		&cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for matching"},
		&cli.BoolFlag{Name: "full-path", Aliases: []string{"p"}, Usage: "Print full paths"},
		&cli.BoolFlag{Name: "unique", Usage: "Print unique subpaths"},
	},
}

var commandRoot = &cli.Command{
	Name:   "root",
	Usage:  "Show repositories' root",
	Action: doRoot,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "all", Usage: "Show all roots"},
	},
}

var commandCreate = &cli.Command{
	Name:   "create",
	Usage:  "Create a new repository",
	Action: doCreate,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend explicitly"},
	},
}

type commandDoc struct {
	Parent    string
	Arguments string
}

var commandDocs = map[string]commandDoc{
	"get":    {"", "[-u] [-p] [--shallow] [--vcs <vcs>] [--look] [--silent] [--brach <branch>] [--no-recursive] <repository URL>|<project>|<user>/<project>|<host>/<user>/<project>"},
	"list":   {"", "[-p] [-e] [<query>]"},
	"create": {"", "<project>|<user>/<project>|<host>/<user>/<project>"},
	"root":   {"", "[-all]"},
}

// Makes template conditionals to generate per-command documents.
func mkCommandsTemplate(genTemplate func(commandDoc) string) string {
	template := "{{if false}}"
	for _, command := range append(commands) {
		template = template + fmt.Sprintf("{{else if (eq .Name %q)}}%s", command.Name, genTemplate(commandDocs[command.Name]))
	}
	return template + "{{end}}"
}

func init() {
	argsTemplate := mkCommandsTemplate(func(doc commandDoc) string { return doc.Arguments })
	parentTemplate := mkCommandsTemplate(func(doc commandDoc) string { return string(strings.TrimLeft(doc.Parent+" ", " ")) })

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
