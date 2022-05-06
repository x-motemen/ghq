= ghq(1) image:https://github.com/x-motemen/ghq/workflows/test/badge.svg?branch=master["Build Status", link="https://github.com/x-motemen/ghq/actions?workflow=test"] image:https://coveralls.io/repos/motemen/ghq/badge.svg?branch=master["Coverage", link="https://coveralls.io/r/motemen/ghq?branch=master"]

== NAME

ghq - Manage remote repository clones

== DESCRIPTION

'ghq' provides a way to organize remote repository clones, like +go get+ does. When you clone a remote repository by +ghq get+, ghq makes a directory under a specific root directory (by default +~/ghq+) using the remote repository URL's host and path.

    $ ghq get https://github.com/x-motemen/ghq
    # Runs `git clone https://github.com/x-motemen/ghq ~/ghq/github.com/x-motemen/ghq`

You can also list local repositories (+ghq list+).

== SYNOPSIS

[verse]
ghq get [-u] [-p] [--shallow] [--vcs <vcs>] [--look] [--silent] [--branch] [--no-recursive] [--bare] <repository URL>|<host>/<user>/<project>|<user>/<project>|<project>
ghq list [-p] [-e] [<query>]
ghq create [--vcs <vcs>] <repository URL>|<host>/<user>/<project>|<user>/<project>|<project>
ghq root [--all]

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
    With '--shallow' option, a "shallow clone" will be performed (for Git
    repositories only, 'git clone --depth 1 ...' eg.). Be careful that a
    shallow-cloned repository cannot be pushed to remote.
    Currently Git and Mercurial repositories are supported. +
    With '--branch' option, you can clone the repository with specified
    repository. This option is currently supported for Git, Mercurial,
    Subversion and git-svn. +
    The 'ghq' gets the git repository recursively by default. +
    We can prevent it with '--no-recursive' option.
    With '--bare' option, a "bare clone" will be performed (for Git
    repositories only, 'git clone --bare ...' eg.).

list::
    List locally cloned repositories. If a query argument is given, only
    repositories whose names contain that query text are listed. '-e'
    ('--exact') forces the match to be an exact one (i.e. the query equals to
    _project_, _user_/_project_ or _host_/_user_/_project_)
    If '-p' ('--full-path') is given, the full paths to the repository root are
    printed instead of relative ones.

root::
    Prints repositories' root (i.e. `ghq.root`). Without '--all' option, the
    primary one is shown.

create::
    Creates new repository.

== CONFIGURATION

Configuration uses 'git-config' variables.

ghq.root::
    The path to directory under which cloned repositories are placed. See
    <<directory-structures,DIRECTORY STRUCTURES>> below. Defaults to +~/ghq+. +
    This variable can have multiple values. If so, the last one becomes
    primary one i.e. new repository clones are always created under it. You may
    want to specify "$GOPATH/src" as a secondary root (environment variables
    should be expanded.)

ghq.<url>.vcs::
    ghq tries to detect the remote repository's VCS backend for non-"github.com"
    repositories.  With this option you can explicitly specify the VCS for the
    remote repository. The URL is matched against '<url>' using 'git config --get-urlmatch'. +
    Accepted values are "git", "github" (an alias for "git"), "subversion",
    "svn" (an alias for "subversion"), "git-svn", "mercurial", "hg" (an alias for "mercurial"),
    "darcs", "fossil", "bazaar", and "bzr" (an alias for "bazaar"). +
    To get this configuration variable effective, you will need Git 1.8.5 or higher.

ghq.<url>.root::
    The "ghq" tries to detect the remote repository-specific root directory. With this option,
    you can specify a repository-specific root directory instead of the common ghq root directory. +
    The URL is matched against '<url>' using 'git config --get-urlmatch'.


=== Example configuration (.gitconfig):

....
[ghq "https://git.example.com/repos/"]
vcs = git
root = ~/myproj
....

== ENVIRONMENT VARIABLES

GHQ_ROOT::
    If set to a path, this value is used as the only root directory regardless
    of other existing ghq.root settings.

== [[directory-structures]]DIRECTORY STRUCTURES

Local repositories are placed under 'ghq.root' with named github.com/_user_/_repo_.

....
~/ghq
|-- code.google.com/
|   `-- p/
|       `-- vim/
`-- github.com/
    |-- google/
    |   `-- go-github/
    |-- motemen/
    |   `-- ghq/
    `-- urfave/
        `-- cli/
....


== [[installing]]INSTALLATION

=== macOS

----
brew install ghq
----

=== Void Linux

----
xbps-install -S ghq
----

=== GNU Guix

----
guix install ghq
----

=== Windows + scoop

----
scoop install ghq
----


=== go get

----
go install github.com/x-motemen/ghq@latest
----

=== conda

----
conda install -c conda-forge go-ghq
----

=== https://github.com/asdf-vm/asdf[asdf-vm]

----
asdf plugin add ghq
asdf install ghq latest
----

=== build

----
git clone https://github.com/x-motemen/ghq .
make install
----

Built binaries are available from GitHub Releases.
https://github.com/x-motemen/ghq/releases

== HANDBOOK

You can buy "ghq-handbook" from Leanpub for more detailed usage.

https://leanpub.com/ghq-handbook

The source Markdown files of this book are also available for free from the following repository.

https://github.com/Songmu/ghq-handbook

Currently, only Japanese version available.
Your translations are welcome!

== AUTHOR

* motemen <motemen@gmail.com>
** https://github.com/sponsors/motemen
* Songmu <y.songmu@gmail.com>
** https://github.com/sponsors/Songmu
