package main

import (
	"os"
	"path/filepath"
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

func TestDoList_query(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		expect string
	}{{
		name:   "repo match",
		args:   []string{"ghq"},
		expect: "github.com/motemen/ghq\n",
	}, {
		name:   "host only doesn't match",
		args:   []string{"github.com"},
		expect: "",
	}, {
		name:   "host and user",
		args:   []string{"github.com/Songmu"},
		expect: "github.com/Songmu/gobump\n",
	}, {
		name:   "with scheme",
		args:   []string{"https://github.com/motemen/ghq"},
		expect: "github.com/motemen/ghq\n",
	}, {
		name:   "exact",
		args:   []string{"-e", "gobump"},
		expect: "github.com/Songmu/gobump\ngithub.com/motemen/gobump\n",
	}, {
		name:   "query",
		args:   []string{"men/go"},
		expect: "github.com/motemen/gobump\ngithub.com/motemen/gore\n",
	}, {
		name:   "exact query",
		args:   []string{"-e", "men/go"},
		expect: "",
	}}

	withFakeGitBackend(t, func(t *testing.T, tmproot string, _ *_cloneArgs, _ *_updateArgs) {
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "ghq", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "gobump", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "motemen", "gore", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmproot, "github.com", "Songmu", "gobump", ".git"), 0755)

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				args := append([]string{"ghq", "list"}, tc.args...)
				out, _, _ := capture(func() {
					newApp().Run(args)
				})
				if out != tc.expect {
					t.Errorf("got:\n%s\nexpect:\n%s", out, tc.expect)
				}
			})
		}
	})

}
