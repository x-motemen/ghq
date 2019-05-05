package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/motemen/ghq/cmdutil"
	"github.com/motemen/ghq/logger"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

var commands = []cli.Command{
	commandGet,
	commandList,
	commandLook,
	commandImport,
	commandRoot,
}

var cloneFlags = []cli.Flag{
	cli.BoolFlag{Name: "update, u", Usage: "Update local repository if cloned already"},
	cli.BoolFlag{Name: "p", Usage: "Clone with SSH"},
	cli.BoolFlag{Name: "shallow", Usage: "Do a shallow clone"},
	cli.BoolFlag{Name: "look, l", Usage: "Look after get"},
	cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for cloning"},
	cli.BoolFlag{Name: "silent, s", Usage: "clone or update silently"},
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
		cli.BoolFlag{Name: "exact, e", Usage: "Perform an exact match"},
		cli.StringFlag{Name: "vcs", Usage: "Specify VCS backend for matching"},
		cli.BoolFlag{Name: "full-path, p", Usage: "Print full paths"},
		cli.BoolFlag{Name: "unique", Usage: "Print unique subpaths"},
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
		cli.BoolFlag{Name: "parallel, P", Usage: "[Experimental] Import parallely"}),
}

var commandRoot = cli.Command{
	Name:   "root",
	Usage:  "Show repositories' root",
	Action: doRoot,
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "all", Usage: "Show all roots"},
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

type getter struct {
	update, shallow, silent, ssh bool
	vcs                          string
}

func (g *getter) get(argURL string) error {
	// If argURL is a "./foo" or "../bar" form,
	// find repository name trailing after github.com/USER/.
	parts := strings.Split(argURL, string(filepath.Separator))
	if parts[0] == "." || parts[0] == ".." {
		if wd, err := os.Getwd(); err == nil {
			path := filepath.Clean(filepath.Join(wd, filepath.Join(parts...)))

			var repoPath string
			roots, err := localRepositoryRoots()
			if err != nil {
				return err
			}
			for _, r := range roots {
				p := strings.TrimPrefix(path, r+string(filepath.Separator))
				if p != path && (repoPath == "" || len(p) < len(repoPath)) {
					repoPath = p
				}
			}

			if repoPath != "" {
				// Guess it
				logger.Log("resolved", fmt.Sprintf("relative %q to %q", argURL, "https://"+repoPath))
				argURL = "https://" + repoPath
			}
		}
	}

	u, err := newURL(argURL)
	if err != nil {
		return xerrors.Errorf("Could not parse URL %q: %w", argURL, err)
	}

	if g.ssh {
		// Assume Git repository if `-p` is given.
		if u, err = convertGitURLHTTPToSSH(u); err != nil {
			return xerrors.Errorf("Could not convet URL %q: %w", u, err)
		}
	}

	remote, err := NewRemoteRepository(u)
	if err != nil {
		return err
	}

	if remote.IsValid() == false {
		return fmt.Errorf("Not a valid repository: %s", u)
	}

	return getRemoteRepository(remote, g.update, g.shallow, g.vcs, g.silent)
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

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func getRemoteRepository(remote RemoteRepository, doUpdate bool, isShallow bool, vcsBackend string, isSilent bool) error {
	remoteURL := remote.URL()
	local, err := LocalRepositoryFromURL(remoteURL)
	if err != nil {
		return err
	}

	path := local.FullPath
	newPath := false

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			newPath = true
			err = nil
		}
		if err != nil {
			return err
		}
	}

	if newPath {
		logger.Log("clone", fmt.Sprintf("%s -> %s", remoteURL, path))

		vcs := vcsRegistry[vcsBackend]
		repoURL := remoteURL
		if vcs == nil {
			vcs, repoURL = remote.VCS()
			if vcs == nil {
				return fmt.Errorf("Could not find version control system: %s", remoteURL)
			}
		}

		err := vcs.Clone(repoURL, path, isShallow, isSilent)
		if err != nil {
			return err
		}
	} else {
		if doUpdate {
			logger.Log("update", path)
			vcs, repoPath := local.VCS()
			if vcs == nil {
				return fmt.Errorf("failed to detect VCS for %q", path)
			}
			vcs.Update(repoPath, isSilent)
		} else {
			logger.Log("exists", path)
		}
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

	reposFound := []*LocalRepository{}
	if err := walkLocalRepositories(func(repo *LocalRepository) {
		if repo.Matches(name) {
			reposFound = append(reposFound, repo)
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
		cmd.Env = append(os.Environ(), "GHQ_LOOK="+repo.RelPath)
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
