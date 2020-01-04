package main

import (
	"os"
	"testing"

	"github.com/Songmu/gitconfig"
)

func TestMain(m *testing.M) {
	teardown := gitconfig.WithConfig(nil, "")
	code := m.Run()
	teardown()
	os.Exit(code)
}
