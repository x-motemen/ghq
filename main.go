package main

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

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
		},
		{
			Name:   "list",
			Usage:  "List local repositories",
			Action: CommandList,
			Flags: []cli.Flag{
				cli.BoolFlag{"exact, e", "Exact match"},
			},
		},
	}

	app.Run(os.Args)
}

func mustBeOkay(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		logInfo("error", fmt.Sprintf("Got unexpected error at %s line %d: %s", file, line, err))
		os.Exit(1)
	}
}

func CommandGet(c *cli.Context) {
	argUrl := c.Args().Get(0)

	if argUrl == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	u, err := ParseGitHubURL(argUrl)
	mustBeOkay(err)

	path := pathForRepository(u)

	newPath := false

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			newPath = true
			err = nil
		}
		mustBeOkay(err)
	}

	if newPath {
		logInfo("clone", fmt.Sprintf("%s -> %s", u, path))

		dir, _ := filepath.Split(path)
		mustBeOkay(os.MkdirAll(dir, 0755))
		Git("clone", u.String(), path)
	} else {
		logInfo("update", path)

		mustBeOkay(os.Chdir(path))
		Git("remote", "update")
	}
}

func CommandList(c *cli.Context) {
	query := c.Args().First()
	exact := c.Bool("exact")

	var filterFn func(string, string, string) bool
	if query == "" {
		filterFn = func(_, _, _ string) bool { return true }
	} else if exact {
		filterFn = func(relPath, user, repo string) bool { return relPath == query || repo == query }
	} else {
		filterFn = func(relPath, user, repo string) bool { return strings.Contains(relPath, query) }
	}

	walkLocalRepositories(func(relPath, user, repo string) {
		if filterFn(relPath, user, repo) == false {
			return
		}

		fmt.Println(relPath)
	})
}

func walkLocalRepositories(callback func(string, string, string)) {
	root := reposRoot()
	filepath.Walk(root, func(path string, fileInfo os.FileInfo, err error) error {
		rel, err := filepath.Rel(root, path)
		mustBeOkay(err)

		user, repo := filepath.Split(rel)
		if user == "" || repo == "" {
			return nil
		}

		callback(rel, user, repo)

		return filepath.SkipDir
	})

	return
}

func reposRoot() string {
	reposRoot, err := GitConfig("ghq.root")
	mustBeOkay(err)

	if reposRoot == "" {
		usr, err := user.Current()
		mustBeOkay(err)

		reposRoot = path.Join(usr.HomeDir, ".ghq", "repos")
	}

	return reposRoot
}

func pathForRepository(u *GitHubURL) string {
	return path.Join(reposRoot(), "@"+u.User, u.Repo)
}

type GitHubURL struct {
	*url.URL
	User string
	Repo string
}

func ParseGitHubURL(urlString string) (*GitHubURL, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	if !u.IsAbs() {
		u.Scheme = "https"
		u.Host = "github.com"
		if u.Path[0] != '/' {
			u.Path = "/" + u.Path
		}
	}

	if u.Host != "github.com" {
		return nil, fmt.Errorf("URL is not of github.com: %s", u)
	}

	components := strings.Split(u.Path, "/")
	if len(components) < 3 {
		return nil, fmt.Errorf("URL does not contain user and repo: %s %v", u, components)
	}
	user, repo := components[1], components[2]

	return &GitHubURL{u, user, repo}, nil
}
