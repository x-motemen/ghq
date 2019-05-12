package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// metaImport represents the parsed <meta name="go-import"
// content="prefix vcs reporoot" /> tags from HTML files.
type metaImport struct {
	Prefix, VCS, RepoRoot string
}

func detectGoImport(u *url.URL) (string, *url.URL, error) {
	goGetU := &url.URL{ // clone
		Scheme:   u.Scheme,
		User:     u.User,
		Host:     u.Host,
		Path:     u.Path,
		RawQuery: u.RawQuery,
	}
	q := goGetU.Query()
	q.Add("go-get", "1")
	goGetU.RawQuery = q.Encode()

	cli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// never follow redirection
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest(http.MethodGet, goGetU.String(), nil)
	req.Header.Add("User-Agent", fmt.Sprintf("ghq/%s (+https://github.com/motemen/ghq)", version))
	resp, err := cli.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	return detectVCSAndRepoURL(resp.Body)
}

// find meta tag like following from thml
// <meta name="go-import" content="gopkg.in/yaml.v2 git https://gopkg.in/yaml.v2">
// ref. https://golang.org/cmd/go/#hdr-Remote_import_paths
func detectVCSAndRepoURL(r io.Reader) (string, *url.URL, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", nil, err
	}

	var mImport *metaImport

	var f func(*html.Node)
	f = func(n *html.Node) {
		if mImport != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "meta" {
			var (
				goImportMeta = false
				content      = ""
			)
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "go-import" {
					goImportMeta = true
					continue
				}
				if a.Key == "content" {
					content = a.Val
				}
			}
			if f := strings.Fields(content); goImportMeta && len(f) == 3 && f[1] != "mod" {
				mImport = &metaImport{
					Prefix:   f[0],
					VCS:      f[1],
					RepoRoot: f[2],
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if mImport == nil {
		return "", nil, fmt.Errorf("no go-import meta tags detected")
	}
	u, err := url.Parse(mImport.RepoRoot)
	if err != nil {
		return "", nil, err
	}
	return mImport.VCS, u, nil
}
