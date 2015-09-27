= ghq(1) image:https://app.wercker.com/status/529f9ef4a8e48e2634661d7f2da9523f/s/master["wercker status", link="https://app.wercker.com/project/bykey/529f9ef4a8e48e2634661d7f2da9523f"]

== NAME

ghq - Manage remote repository clones

== DESCRIPTION

'ghq' provides a way to organize remote repository clones, like +go get+ does. When you clone a remote repository by +ghq get+, ghq makes a directory under a specific root directory (by default +~/.ghq+) using the remote repository URL's host and path.

    $ ghq get https://github.com/motemen/ghq
    # Runs `git clone https://github.com/motemen/ghq ~/.ghq/github.com/motemen/ghq`

You can also list local repositories (+ghq list+), jump into local repositories (+ghq look+), and bulk get repositories by list of URLs (+ghq import+).

== SYNOPSIS

[verse]
'ghq' get [-u] [-p] (<repository URL> | <user>/<project> | <project>)
'ghq' list [-p] [-e] [<query>]
'ghq' look (<project> | <path/to/project>)
'ghq' import [-u] [-p] < FILE
'ghq' import <subcommand> [<args>...]
'ghq' root [--all]

== COMMANDS

get::
    Clone a remote repository under ghq root directory (see
    <<directory-structures,DIRECTORY STRUCTURES>> below). If the repository is
    already cloned to local, nothing will happen unless '-u' ('--update')
    flag is supplied, in which case the local repository is updated ('git pull --ff-only' eg.).
    When you use '-p' option, the repository is cloned via SSH protocol. +
    If there are multiple +ghq.root+ s, existing local clones are searched
    first. Then a new repository clone is created under the primary root if
    none is found. +
    With '-shallow' option, a "shallow clone" will be performed (for Git
    repositories only, 'git clone --depth 1 ...' eg.). Be careful that a
    shallow-cloned repository cannot be pushed to remote. +
    Currently Git and Mercurial repositories are supported.

list::
    List locally cloned repositories. If a query argument is given, only
    repositories whose names contain that query text are listed. '-e'
    ('--exact') forces the match to be an exact one (i.e. the query equals to
    _project_ or _user_/_project_) If '-p' ('--full-path') is given, the full paths
    to the repository root are printed instead of relative ones.

look::
    Look into a locally cloned repository with the shell.

import::
    If no extra arguments given, reads repository URLs from stdin line by line
    and performs 'get' for each of them. +
    If given a subcommand name e.g. 'ghq import <subcommand> [<args>...]',
    ghq looks up a configuration 'ghq.import.<subcommand>' for a command, invokes
    it, and uses its output as URLs list. See below for 'ghq.import.<subcommand>'
    in CONFIGURATION section.

root::
    Prints repositories' root (i.e. `ghq.root`). Without '--all' option, the
    primary one is shown.

== CONFIGURATION

Configuration uses 'git-config' variables.

ghq.root::
    The path to directory under which cloned repositories are placed. See
    <<directory-structures,DIRECTORY STRUCTURES>> below. Defaults to +~/.ghq+. +
    This variable can have multiple values. If so, the first one becomes
    primary one i.e. new repository clones are always created under it. You may
    want to specify "$GOPATH/src" as a secondary root (environment variables
    should be expanded.)

ghq.<url>.vcs::
    ghq tries to detect the remote repository's VCS backend for non-"github.com"
    repositories.  With this option you can explicitly specify the VCS for the
    remote repository. The URL is matched against '<url>' using 'git config --get-urlmatch'. +
    Accepted values are "git", "github" (an alias for "git"), "subversion",
    "svn" (an alias for "subversion"), "git-svn", "mercurial", "hg" (an alias for "mercurial"),
    and "darcs". +
    To get this configuration variable effective, you will need Git 1.8.5 or higher. +
    For example in .gitconfig:

....
[ghq "https://git.example.com/repos/"]
vcs = git
....


ghq.import.<subcommand>::
    When 'import' is called with extra arguments e.g. 'ghq import <subcommand> [<args>...]',
    first of them is treated as a subcommand name and this configuration value
    will be used for a command. The command is invoked with rest arguments
    and expected to print remote repository URLs line by line. +
    For example with https://github.com/motemen/github-list-starred[github-list-starred]:

....
# Invoke as `ghq import starred motemen`
[ghq "import"]
starred = github-list-starred
....


ghq.ghe.host::
    The hostname of your GitHub Enterprise installation. A repository that has a
    hostname set with this key will be regarded as same one as one on GitHub.
    This variable can have multiple values. If so, `ghq` tries matching with
    each hostnames. +
    This option is DEPRECATED, so use "ghq.<url>.vcs" configuration instead.

== ENVIRONMENT VARIABLES

GHQ_ROOT::
    If set to a path, this value is used as the only root directory regardless
    of other existing ghq.root settings.

== [[directory-structures]]DIRECTORY STRUCTURES

Local repositories are placed under 'ghq.root' with named github.com/_user_/_repo_.

....
~/.ghq
|-- code.google.com/
|   `-- p/
|       `-- vim/
`-- github.com/
    |-- codegangsta/
    |   `-- cli/
    |-- google/
    |   `-- go-github/
    `-- motemen/
        `-- ghq/
....


== [[installing]]INSTALLATION

----
go get github.com/motemen/ghq
----

Or clone the https://github.com/motemen/ghq[repository] and run:

----
make install
----

== AUTHOR

motemen <motemen@gmail.com>
