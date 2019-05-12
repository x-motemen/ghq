package main

import (
	"strings"
	"testing"
)

func TestDetectVCSAndRepoURL(t *testing.T) {
	input := `<html>
<head>
<meta name="go-import" content="gopkg.in/yaml.v2 mod https://gopkg.in/yaml.v2">
<meta name="go-import" content="gopkg.in/yaml.v2 git https://gopkg.in/yaml.v2">
<meta name="go-source" content="gopkg.in/yaml.v2 _ https://github.com/go-yaml/yaml/tree/v2.2.2{/dir} https://github.com/go-yaml/yaml/blob/v2.2.2{/dir}/{file}#L{line}">
</head>
<body>
go get gopkg.in/yaml.v2
</body>
</html>`

	vcs, u, err := detectVCSAndRepoURL(strings.NewReader(input))
	if vcs != "git" {
		t.Errorf("want: %q, got: %q", "git", vcs)
	}
	expectedURL := "https://gopkg.in/yaml.v2"
	if u.String() != expectedURL {
		t.Errorf("want: %q, got: %q", expectedURL, u.String())
	}
	if err != nil {
		t.Errorf("something went wrong: %s", err)
	}
}
