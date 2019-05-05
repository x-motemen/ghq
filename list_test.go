package main

import (
	"testing"

	"github.com/urfave/cli"
)

func TestCommandList(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
	}
}

func TestCommandListUnique(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unique"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
	}
}

func TestCommandListUnknown(t *testing.T) {
	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unknown-flag"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	if err != nil {
		t.Errorf("error should be nil, but: %s", err)
	}
}
