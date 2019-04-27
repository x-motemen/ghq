package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/motemen/ghq/logger"
	"github.com/urfave/cli"
)

var Commands = []cli.Command{
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
	Flags:  cloneFlags,
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
	for _, command := range append(Commands) {
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
	argURL := c.Args().Get(0)
	doUpdate := c.Bool("update")
	isShallow := c.Bool("shallow")
	andLook := c.Bool("look")
	vcsBackend := c.String("vcs")

	if argURL == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	// If argURL is a "./foo" or "../bar" form,
	// find repository name trailing after github.com/USER/.
	parts := strings.Split(argURL, string(filepath.Separator))
	if parts[0] == "." || parts[0] == ".." {
		if wd, err := os.Getwd(); err == nil {
			path := filepath.Clean(filepath.Join(wd, filepath.Join(parts...)))

			var repoPath string
			for _, r := range localRepositoryRoots() {
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

	url, err := NewURL(argURL)
	if err != nil {
		return err
	}

	isSSH := c.Bool("p")
	if isSSH {
		// Assume Git repository if `-p` is given.
		if url, err = ConvertGitURLHTTPToSSH(url); err != nil {
			return err
		}
	}

	remote, err := NewRemoteRepository(url)
	if err != nil {
		return err
	}

	if remote.IsValid() == false {
		return fmt.Errorf("Not a valid repository: %s", url)
	}

	if err := getRemoteRepository(remote, doUpdate, isShallow, vcsBackend); err != nil {
		return err
	}
	if andLook {
		doLook(c)
	}
	return nil
}

// getRemoteRepository clones or updates a remote repository remote.
// If doUpdate is true, updates the locally cloned repository. Otherwise does nothing.
// If isShallow is true, does shallow cloning. (no effect if already cloned or the VCS is Mercurial and git-svn)
func getRemoteRepository(remote RemoteRepository, doUpdate bool, isShallow bool, vcsBackend string) error {
	remoteURL := remote.URL()
	local := LocalRepositoryFromURL(remoteURL)

	path := local.FullPath
	newPath := false

	_, err := os.Stat(path)
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

		err := vcs.Clone(repoURL, path, isShallow)
		if err != nil {
			return err
		}
	} else {
		if doUpdate {
			logger.Log("update", path)
			local.VCS().Update(path)
		} else {
			logger.Log("exists", path)
		}
	}
	return nil
}

func doList(c *cli.Context) error {
	query := c.Args().First()
	exact := c.Bool("exact")
	printFullPaths := c.Bool("full-path")
	printUniquePaths := c.Bool("unique")

	var filterFn func(*LocalRepository) bool
	if query == "" {
		filterFn = func(_ *LocalRepository) bool {
			return true
		}
	} else if exact {
		filterFn = func(repo *LocalRepository) bool {
			return repo.Matches(query)
		}
	} else {
		filterFn = func(repo *LocalRepository) bool {
			return strings.Contains(repo.NonHostPath(), query)
		}
	}

	repos := []*LocalRepository{}

	walkLocalRepositories(func(repo *LocalRepository) {
		if filterFn(repo) == false {
			return
		}

		repos = append(repos, repo)
	})

	if printUniquePaths {
		subpathCount := map[string]int{} // Count duplicated subpaths (ex. foo/dotfiles and bar/dotfiles)
		reposCount := map[string]int{}   // Check duplicated repositories among roots

		// Primary first
		for _, repo := range repos {
			if reposCount[repo.RelPath] == 0 {
				for _, p := range repo.Subpaths() {
					subpathCount[p] = subpathCount[p] + 1
				}
			}

			reposCount[repo.RelPath] = reposCount[repo.RelPath] + 1
		}

		for _, repo := range repos {
			if reposCount[repo.RelPath] > 1 && repo.IsUnderPrimaryRoot() == false {
				continue
			}

			for _, p := range repo.Subpaths() {
				if subpathCount[p] == 1 {
					fmt.Println(p)
					break
				}
			}
		}
	} else {
		for _, repo := range repos {
			if printFullPaths {
				fmt.Println(repo.FullPath)
			} else {
				fmt.Println(repo.RelPath)
			}
		}
	}
	return nil
}

func doLook(c *cli.Context) error {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "look")
		os.Exit(1)
	}

	reposFound := []*LocalRepository{}
	walkLocalRepositories(func(repo *LocalRepository) {
		if repo.Matches(name) {
			reposFound = append(reposFound, repo)
		}
	})

	if len(reposFound) == 0 {
		url, err := NewURL(name)

		if err == nil {
			repo := LocalRepositoryFromURL(url)
			_, err := os.Stat(repo.FullPath)

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
		if runtime.GOOS == "windows" {
			cmd := exec.Command(os.Getenv("COMSPEC"))
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = reposFound[0].FullPath
			err := cmd.Start()
			if err == nil {
				cmd.Wait()
				return nil
			}
		} else {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}

			logger.Log("cd", reposFound[0].FullPath)
			if err := os.Chdir(reposFound[0].FullPath); err != nil {
				return err
			}
			env := append(syscall.Environ(), "GHQ_LOOK="+reposFound[0].RelPath)
			syscall.Exec(shell, []string{shell}, env)
		}

	default:
		logger.Log("error", "More than one repositories are found; Try more precise name")
		for _, repo := range reposFound {
			logger.Log("error", "- "+strings.Join(repo.PathParts, "/"))
		}
	}
	return nil
}

func doImport(c *cli.Context) error {
	var (
		doUpdate   = c.Bool("update")
		isSSH      = c.Bool("p")
		isShallow  = c.Bool("shallow")
		vcsBackend = c.String("vcs")
	)

	var (
		in       io.Reader
		finalize func() error
	)

	if len(c.Args()) == 0 {
		// `ghq import` reads URLs from stdin
		in = os.Stdin
		finalize = func() error { return nil }
	} else {
		// Handle `ghq import starred motemen` case
		// with `git config --global ghq.import.starred "!github-list-starred"`
		subCommand := c.Args().First()
		command, err := GitConfigSingle("ghq.import." + subCommand)
		if err == nil && command == "" {
			err = fmt.Errorf("ghq.import.%s configuration not found", subCommand)
		}
		if err != nil {
			return err
		}

		// execute `sh -c 'COMMAND "$@"' -- ARG...`
		// TODO: Windows
		command = strings.TrimLeft(command, "!")
		shellCommand := append([]string{"sh", "-c", command + ` "$@"`, "--"}, c.Args().Tail()...)

		logger.Log("run", strings.Join(append([]string{command}, c.Args().Tail()...), " "))

		cmd := exec.Command(shellCommand[0], shellCommand[1:]...)
		cmd.Stderr = os.Stderr

		in, err = cmd.StdoutPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		finalize = cmd.Wait
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		url, err := NewURL(line)
		if err != nil {
			logger.Log("error", fmt.Sprintf("Could not parse URL <%s>: %s", line, err))
			continue
		}
		if isSSH {
			url, err = ConvertGitURLHTTPToSSH(url)
			if err != nil {
				logger.Log("error", fmt.Sprintf("Could not convert URL <%s>: %s", url, err))
				continue
			}
		}

		remote, err := NewRemoteRepository(url)
		if logger.ErrorIf(err) {
			continue
		}
		if remote.IsValid() == false {
			logger.Log("error", fmt.Sprintf("Not a valid repository: %s", url))
			continue
		}

		if err := getRemoteRepository(remote, doUpdate, isShallow, vcsBackend); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("While reading input: %s", err)
	}

	return finalize()
}

func doRoot(c *cli.Context) error {
	all := c.Bool("all")
	if all {
		for _, root := range localRepositoryRoots() {
			fmt.Println(root)
		}
	} else {
		fmt.Println(primaryLocalRepositoryRoot())
	}
	return nil
}
