package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/motemen/ghq/cmdutil"
	"github.com/urfave/cli"
)

var commands = []cli.Command{
	commandGet,
	commandList,
	commandLook,
	commandImport,
	commandRoot,
}

var cloneFlags = []cli.Flag{
	&cli.BoolFlag{Name: "update, u", Usage: "Update local repository if cloned already"},
	&cli.BoolFlag{Name: "p", Usage: "Clone with SSH"},
	&cli.BoolFlag{Name: "shallow", Usage: "Do a shallow clone"},
	&cli.BoolFlag{Name: "look, l", Usage: "Look after get"},
	&cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for cloning"},
	&cli.BoolFlag{Name: "silent, s", Usage: "clone or update silently"},
	&cli.StringFlag{Name: "branch, b", Usage: "Specify branch name. This flag implies --single-branch on Git"},
}

var commandGet = cli.Command{
	Name:  "get",
	Usage: "Clone/sync with a remote repository",
	Description: `
    Clone a GitHub repository under ghq root directory. If the repository is
    already cloned to local, nothing will happen unless '-u' ('--update')
    flag is supplied, in which case 'git remote update' is executed.
    When you use '-p' option, the repository is cloned via SSH.
`,
	Action: doGet,
	Flags:  cloneFlags,
}

var commandList = cli.Command{
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
		&cli.BoolFlag{Name: "exact, e", Usage: "Perform an exact match"},
		&cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for matching"},
		&cli.BoolFlag{Name: "full-path, p", Usage: "Print full paths"},
		&cli.BoolFlag{Name: "unique", Usage: "Print unique subpaths"},
	},
}

var commandLook = cli.Command{
	Name:  "look",
	Usage: "Look into a local repository",
	Description: `
    Look into a locally cloned repository with the shell.
`,
	Action: doLook,
}

var commandImport = cli.Command{
	Name:   "import",
	Usage:  "Bulk get repositories from stdin",
	Action: doImport,
	Flags: append(cloneFlags,
		&cli.BoolFlag{Name: "parallel, P", Usage: "[Experimental] Import parallely"}),
}

var commandRoot = cli.Command{
	Name:   "root",
	Usage:  "Show repositories' root",
	Action: doRoot,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "all", Usage: "Show all roots"},
	},
}

type commandDoc struct {
	Parent    string
	Arguments string
}

var commandDocs = map[string]commandDoc{
	"get":    {"", "[-u] [--vcs <vcs>] <repository URL> | [-u] [-p] <user>/<project>"},
	"list":   {"", "[-p] [-e] [<query>]"},
	"look":   {"", "<project> | <user>/<project> | <host>/<user>/<project>"},
	"import": {"", "< file"},
	"root":   {"", ""},
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

func doGet(c *cli.Context) error {
	var (
		argURL  = c.Args().Get(0)
		andLook = c.Bool("look")
	)
	g := &getter{
		update:  c.Bool("update"),
		shallow: c.Bool("shallow"),
		ssh:     c.Bool("p"),
		vcs:     c.String("vcs"),
		silent:  c.Bool("silent"),
		branch:  c.String("branch"),
	}

	if argURL == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	if err := g.get(argURL); err != nil {
		return err
	}
	if andLook {
		return doLook(c)
	}
	return nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return shell
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("COMSPEC")
	}
	return "/bin/sh"
}

func doLook(c *cli.Context) error {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "look")
		os.Exit(1)
	}

	var (
		reposFound []*LocalRepository
		mu         sync.Mutex
	)
	if err := walkAllLocalRepositories(func(repo *LocalRepository) {
		if repo.Matches(name) {
			mu.Lock()
			reposFound = append(reposFound, repo)
			mu.Unlock()
		}
	}); err != nil {
		return err
	}

	if len(reposFound) == 0 {
		if url, err := newURL(name); err == nil {
			repo, err := LocalRepositoryFromURL(url)
			if err != nil {
				return err
			}
			_, err = os.Stat(repo.FullPath)

			// if the directory exists
			if err == nil {
				reposFound = append(reposFound, repo)
			}
		}
	}

	switch len(reposFound) {
	case 0:
		return fmt.Errorf("No repository found")
	case 1:
		repo := reposFound[0]
		cmd := exec.Command(detectShell())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = repo.FullPath
		cmd.Env = append(os.Environ(), "GHQ_LOOK="+filepath.ToSlash(repo.RelPath))
		return cmdutil.RunCommand(cmd, true)
	default:
		b := &strings.Builder{}
		b.WriteString("More than one repositories are found; Try more precise name\n")
		for _, repo := range reposFound {
			b.WriteString(fmt.Sprintf("       - %s\n", strings.Join(repo.PathParts, "/")))
		}
		return errors.New(b.String())
	}
}

func doRoot(c *cli.Context) error {
	var (
		w   = c.App.Writer
		all = c.Bool("all")
	)
	if all {
		roots, err := localRepositoryRoots()
		if err != nil {
			return err
		}
		for _, root := range roots {
			fmt.Fprintln(w, root)
		}
		return nil
	}
	root, err := primaryLocalRepositoryRoot()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, root)
	return nil
}
