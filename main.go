package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/motemen/ghq/pocket"
	"github.com/motemen/ghq/utils"
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
				cli.BoolFlag{"full-path, p", "Show full paths"},
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

func mustBeOkay(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		utils.Log("error", fmt.Sprintf("at %s line %d: %s", file, line, err))
		os.Exit(1)
	}
}

func CommandGet(c *cli.Context) {
	argUrl := c.Args().Get(0)
	doUpdate := c.Bool("update")

	if argUrl == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	u, err := ParseGitHubURL(argUrl)
	mustBeOkay(err)

	GetGitHubRepository(u, doUpdate)
}

func GetGitHubRepository(u *GitHubURL, doUpdate bool) {
	repo := LocalRepositoryFromGitHubURL(u)

	path := repo.FullPath
	newPath := false

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			newPath = true
			err = nil
		}
		mustBeOkay(err)
	}

	if newPath {
		utils.Log("clone", fmt.Sprintf("%s/%s -> %s", u.User, u.Repo, path))

		dir, _ := filepath.Split(path)
		mustBeOkay(os.MkdirAll(dir, 0755))
		Git("clone", u.String(), path)
	} else {
		if doUpdate {
			utils.Log("update", path)

			mustBeOkay(os.Chdir(path))
			Git("remote", "update")
		} else {
			utils.Log("exists", path)
		}
	}
}

func CommandList(c *cli.Context) {
	query := c.Args().First()
	exact := c.Bool("exact")
	showFullPath := c.Bool("full-path")

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

	walkLocalRepositories(func(repo *LocalRepository) {
		if filterFn(repo) == false {
			return
		}

		if showFullPath {
			fmt.Println(repo.FullPath)
		} else {
			fmt.Println(repo.RelPath)
		}
	})
}

func CommandLook(c *cli.Context) {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "look")
		os.Exit(1)
	}

	reposFound := make([]*LocalRepository, 0)
	walkLocalRepositories(func(repo *LocalRepository) {
		if repo.Matches(name) {
			reposFound = append(reposFound, repo)
		}
	})

	switch len(reposFound) {
	case 0:
		utils.Log("error", "No repository found")

	case 1:
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		mustBeOkay(os.Chdir(reposFound[0].FullPath))
		syscall.Exec(shell, []string{shell}, syscall.Environ())

	default:
		utils.Log("error", "More than one repositories are found; Try more precise name")
		for _, repo := range reposFound {
			utils.Log("error", "- "+strings.Join(repo.PathParts, "/"))
		}
	}
}

func CommandPocket(c *cli.Context) {
	accessToken, err := GitConfig("ghq.pocket.token")
	mustBeOkay(err)

	if accessToken == "" {
		receiverURL, ch, err := pocket.StartAccessTokenReceiver()
		mustBeOkay(err)
		utils.Log("pocket", "Waiting for Pocket authentication callback at "+receiverURL)

		utils.Log("pocket", "Obtaining request token")
		authRequest, err := pocket.ObtainRequestToken(receiverURL)
		mustBeOkay(err)

		url := pocket.GenerateAuthorizationURL(authRequest.Code, receiverURL)
		utils.Log("open", url)

		<-ch

		utils.Log("pocket", "Obtaining access token")
		authorized, err := pocket.ObtainAccessToken(authRequest.Code)
		mustBeOkay(err)

		utils.Log("authorized", authorized.Username)

		accessToken = authorized.AccessToken
		Git("config", "ghq.pocket.token", authorized.AccessToken)
	}

	utils.Log("pocket", "Retrieving github.com entries")
	res, err := pocket.RetrieveGitHubEntries(accessToken)
	mustBeOkay(err)

	for _, item := range res.List {
		u, err := ParseGitHubURL(item.ResolvedURL)
		if err != nil {
			utils.Log("error", err.Error())
			continue
		} else if u.User == "blog" {
			utils.Log("skip", fmt.Sprintf("%s: is not a repository", u))
			continue
		} else if u.Extra != "" {
			utils.Log("skip", fmt.Sprintf("%s: is not project home", u))
			continue
		}

		GetGitHubRepository(u, c.Bool("update"))
	}
}
