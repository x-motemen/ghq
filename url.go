package main

import (
	"fmt"
	"net/url"
	"regexp"
)

var pattern = regexp.MustCompile("^([^@]+)@([^:]+):(.+)$")

func NewURL(ref string) (*url.URL, error) {
	if pattern.MatchString(ref) {
		matched := pattern.FindStringSubmatch(ref)
		ref = fmt.Sprintf("ssh://%s@%s/%s", matched[1], matched[2], matched[3])
	}

	return url.Parse(ref)
}
