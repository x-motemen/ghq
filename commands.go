package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/motemen/ghq/pocket"
	"github.com/motemen/ghq/utils"
)

func CommandGet(c *cli.Context) {
	argUrl := c.Args().Get(0)
	doUpdate := c.Bool("update")

	if argUrl == "" {
		cli.ShowCommandHelp(c, "get")
		os.Exit(1)
	}

	url, err := url.Parse(argUrl)
	mustBeOkay(err)

	if !url.IsAbs() {
		url.Scheme = "https"
		url.Host = "github.com"
		if url.Path[0] != '/' {
			url.Path = "/" + url.Path
		}
	}

	remote, err := NewRemoteRepository(url)
	mustBeOkay(err)

	if remote.IsValid() == false {
		utils.Log("error", fmt.Sprintf("Not a valid repository: %s", url))
		os.Exit(1)
	}

	getRemoteRepository(remote, doUpdate)
}

func getRemoteRepository(remote RemoteRepository, doUpdate bool) {
	remoteURL := remote.RepositoryURL()
	pathParts := append(
		[]string{remoteURL.Host}, strings.Split(remote.RepositoryURL().Path, "/")...,
	)
	local := LocalRepositoryFromPathParts(pathParts)

	path := local.FullPath
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
		utils.Log("clone", fmt.Sprintf("%s -> %s", remote.RepositoryURL(), path))

		remote.VCS().Clone(remote.RepositoryURL(), path)
	} else {
		if doUpdate {
			utils.Log("update", path)

			remote.VCS().Update(path)
		} else {
			utils.Log("exists", path)
		}
	}
}

func CommandList(c *cli.Context) {
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

	repos := make([]*LocalRepository, 0)

	walkLocalRepositories(func(repo *LocalRepository) {
		if filterFn(repo) == false {
			return
		}

		repos = append(repos, repo)
	})

	if printUniquePaths {
		subpathCount := make(map[string]int)

		for _, repo := range repos {
			for _, p := range repo.Subpaths() {
				subpathCount[p] = subpathCount[p] + 1
			}
		}

		for _, repo := range repos {
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
		url, err := url.Parse(item.ResolvedURL)
		if err != nil {
			utils.Log("error", fmt.Sprintf("Could not parse URL <%s>: %s", item.ResolvedURL, err))
			continue
		}

		remote, err := NewRemoteRepository(url)
		mustBeOkay(err)

		if remote.IsValid() == false {
			utils.Log("error", fmt.Sprintf("Not a valid repository: %s", url))
			continue
		}

		getRemoteRepository(remote, c.Bool("update"))
	}
}
