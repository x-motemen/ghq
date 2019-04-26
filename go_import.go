package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func detectGoImport(u *url.URL) (string, *url.URL, error) {
	goGetU, _ := url.Parse(u.String()) // clone
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
	req.Header.Add("User-Agent", fmt.Sprintf("ghq/%s (+https://github.com/motemen/ghq)", Version))
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

	var goImportContent string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if goImportContent != "" {
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
			if goImportMeta && content != "" {
				goImportContent = content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	stuffs := strings.Fields(goImportContent)
	if len(stuffs) < 3 {
		return "", nil, fmt.Errorf("no go-import meta tags detected")
	}
	u, err := url.Parse(stuffs[2])
	if err != nil {
		return "", nil, err
	}
	return stuffs[1], u, nil
}
