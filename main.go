package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

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
		log.Fatalf("Got unexpected error at %s line %d: %s", file, line, err)
	}
}

func Git(command ...string) {
	log.Printf("Running 'git %s'\n", strings.Join(command, " "))
	cmd := exec.Command("git", command...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("git %s: %s", strings.Join(command, " "), err)
	}
}

func GitConfig(key string) (string, error) {
	defaultValue := ""

	cmd := exec.Command("git", "config", "--path", "--null", "--get", key)
	cmd.Stderr = os.Stderr

	buf, err := cmd.Output()

	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() == 1 {
				return defaultValue, nil
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return strings.TrimRight(string(buf), "\000"), nil
}

func CommandGet(c *cli.Context) {
	argUrl := c.Args().Get(0)

	if argUrl == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	u, err := ParseGitHubURL(argUrl)
	if err != nil {
		log.Fatalf("While parsing URL: %s", err)
	}

	path := pathForRepository(u)
	if err != nil {
		log.Fatalf("Could not obtain path for repository %s: %s", u, err)
	}

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
		dir, _ := filepath.Split(path)
		mustBeOkay(os.MkdirAll(dir, 0755))
		Git("clone", u.String(), path)
	} else {
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
		filterFn = func(path, user, repo string) bool { return path == query || repo == query }
	} else {
		filterFn = func(path, user, repo string) bool { return strings.Contains(path, query) }
	}

	walkLocalRepositories(func(relPath string, user string, repo string) {
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
