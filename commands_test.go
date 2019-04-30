package main

import (
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/urfave/cli"
)

func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func captureReader(block func()) (*os.File, *os.File, error) {
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	rErr, wErr, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	defer wOut.Close()
	defer wErr.Close()

	var stdout, stderr *os.File
	os.Stdout, stdout = wOut, os.Stdout
	os.Stderr, stderr = wErr, os.Stderr

	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	block()

	wOut.Close()
	wErr.Close()

	return rOut, rErr, nil
}

func capture(block func()) (string, string, error) {
	rOut, rErr, err := captureReader(block)
	if err != nil {
		return "", "", err
	}

	bufOut, err := ioutil.ReadAll(rOut)
	if err != nil {
		return "", "", err
	}

	bufErr, err := ioutil.ReadAll(rErr)
	if err != nil {
		return "", "", err
	}

	return string(bufOut), string(bufErr), nil
}

type _cloneArgs struct {
	remote  *url.URL
	local   string
	shallow bool
}

type _updateArgs struct {
	local string
}

func withFakeGitBackend(t *testing.T, block func(string, *_cloneArgs, *_updateArgs)) {
	tmpRoot, err := ioutil.TempDir("", "ghq")
	if err != nil {
		t.Fatalf("Could not create tempdir: %s", err)
	}
	defer os.RemoveAll(tmpRoot)

	// Resolve /var/folders/.../T/... to /private/var/... in OSX
	tmpRoot = func() string {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd(): %s", err)
		}

		defer func() {
			err := os.Chdir(wd)
			if err != nil {
				t.Fatalf("os.Chdir(): %s", err)
			}
		}()

		err = os.Chdir(tmpRoot)
		if err != nil {
			t.Fatalf("os.Chdir(): %s", err)
		}

		tmpRoot, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd(): %s", err)
		}

		return tmpRoot
	}()

	_localRepositoryRoots = []string{tmpRoot}
	defer func() { _localRepositoryRoots = []string{} }()

	var cloneArgs _cloneArgs
	var updateArgs _updateArgs

	var originalGitBackend = GitBackend
	tmpBackend := &VCSBackend{
		Clone: func(remote *url.URL, local string, shallow, silent bool) error {
			cloneArgs = _cloneArgs{
				remote:  remote,
				local:   filepath.FromSlash(local),
				shallow: shallow,
			}
			return nil
		},
		Update: func(local string, silent bool) error {
			updateArgs = _updateArgs{
				local: local,
			}
			return nil
		},
	}
	GitBackend = tmpBackend
	vcsDirsMap[".git"] = tmpBackend
	defer func() { GitBackend = originalGitBackend; vcsDirsMap[".git"] = originalGitBackend }()

	block(tmpRoot, &cloneArgs, &updateArgs)
}

func TestCommandGet(t *testing.T) {
	RegisterTestingT(t)
	app := newApp()

	testCases := []struct {
		name string
		f    func(string, *_cloneArgs, *_updateArgs)
	}{
		{
			name: "simple",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "motemen/ghq-test-repo"})

				Expect(cloneArgs.remote.String()).To(Equal("https://github.com/motemen/ghq-test-repo"))
				Expect(filepath.ToSlash(cloneArgs.local)).To(Equal(filepath.ToSlash(localDir)))
				Expect(cloneArgs.shallow).To(Equal(false))
			},
		},
		{
			name: "p option",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "-p", "motemen/ghq-test-repo"})

				Expect(cloneArgs.remote.String()).To(Equal("ssh://git@github.com/motemen/ghq-test-repo"))
				Expect(filepath.ToSlash(cloneArgs.local)).To(Equal(filepath.ToSlash(localDir)))
				Expect(cloneArgs.shallow).To(Equal(false))
			},
		},
		{
			name: "already cloned with -u",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
				// mark as "already cloned", the condition may change later
				os.MkdirAll(filepath.Join(localDir, ".git"), 0755)

				app.Run([]string{"", "get", "-u", "motemen/ghq-test-repo"})

				Expect(updateArgs.local).To(Equal(localDir))
			},
		},
		{
			name: "shallow",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")

				app.Run([]string{"", "get", "-shallow", "motemen/ghq-test-repo"})

				Expect(cloneArgs.remote.String()).To(Equal("https://github.com/motemen/ghq-test-repo"))
				Expect(filepath.ToSlash(cloneArgs.local)).To(Equal(filepath.ToSlash(localDir)))
				Expect(cloneArgs.shallow).To(Equal(true))
			},
		},
		{
			name: "dot slach ./",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen")
				os.MkdirAll(localDir, 0755)
				wd, _ := os.Getwd()
				defer os.Chdir(wd)
				os.Chdir(localDir)

				app.Run([]string{"", "get", "-u", "." + string(filepath.Separator) + "ghq-test-repo"})

				Expect(cloneArgs.remote.String()).To(Equal("https://github.com/motemen/ghq-test-repo"))
				Expect(cloneArgs.local).To(Equal(filepath.Join(localDir, "ghq-test-repo")))
			},
		},
		{
			name: "dot dot slash ../",
			f: func(tmpRoot string, cloneArgs *_cloneArgs, updateArgs *_updateArgs) {
				localDir := filepath.Join(tmpRoot, "github.com", "motemen", "ghq-test-repo")
				os.MkdirAll(localDir, 0755)
				wd, _ := os.Getwd()
				defer os.Chdir(wd)
				os.Chdir(localDir)

				app.Run([]string{"", "get", "-u", ".." + string(filepath.Separator) + "ghq-another-test-repo"})

				Expect(cloneArgs.remote.String()).To(Equal("https://github.com/motemen/ghq-another-test-repo"))
				Expect(cloneArgs.local).To(Equal(filepath.Join(tmpRoot, "github.com", "motemen", "ghq-another-test-repo")))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withFakeGitBackend(t, tc.f)
		})
	}
}

func TestCommandList(t *testing.T) {
	RegisterTestingT(t)

	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	Expect(err).To(BeNil())
}

func TestCommandListUnique(t *testing.T) {
	RegisterTestingT(t)

	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unique"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	Expect(err).To(BeNil())
}

func TestCommandListUnknown(t *testing.T) {
	RegisterTestingT(t)

	_, _, err := capture(func() {
		app := cli.NewApp()
		flagSet := flagSet("list", commandList.Flags)
		flagSet.Parse([]string{"--unknown-flag"})
		c := cli.NewContext(app, flagSet, nil)

		doList(c)
	})

	Expect(err).To(BeNil())
}
