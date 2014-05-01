package main

import (
	"fmt"
	"net/url"
	"strings"
)

type GitHubURL struct {
	*url.URL
	User  string
	Repo  string
	Extra string
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

	components := strings.SplitN(u.Path, "/", 4)
	if len(components) < 3 {
		return nil, fmt.Errorf("URL does not contain user and repo: %s %v", u, components)
	}

	gu := &GitHubURL{URL: u}
	gu.User, gu.Repo = components[1], components[2]
	if len(components) > 3 {
		gu.Extra = components[3]
	}

	return gu, nil
}
