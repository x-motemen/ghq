#!/bin/sh
set -e

tmpdir=$(mktemp -d)

cleanup() {
    code=$?
    rm -rf $tmpdir
    exit $code
}
trap cleanup EXIT

set -x

export GHQ_ROOT=$tmpdir

ghq get motemen/ghq
ghq get www.mercurial-scm.org/repo/hello
ghq get https://launchpad.net/terminator
ghq get --vcs fossil https://www.sqlite.org/src
ghq get --shallow --vcs=git-svn https://svn.apache.org/repos/asf/httpd/httpd
ghq get https://svn.apache.org/repos/asf/subversion

test -d $tmpdir/github.com/motemen/ghq/.git
test -d $tmpdir/www.mercurial-scm.org/repo/hello/.hg
test -d $tmpdir/launchpad.net/terminator/.bzr
test -f $tmpdir/www.sqlite.org/src/.fslckout
test -d $tmpdir/svn.apache.org/repos/asf/httpd/httpd/.git/svn
test -d $tmpdir/svn.apache.org/repos/asf/subversion/.svn
