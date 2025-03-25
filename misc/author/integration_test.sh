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

: testing 'ghq get'
    ghq get x-motemen/ghq
    ghq get www.mercurial-scm.org/repo/hello
    ghq get https://launchpad.net/shutter
    ghq get --vcs fossil https://www.sqlite.org/src
    ghq get http://chiselapp.com/user/sti/repository/fossil-gui
    ghq get --shallow --vcs=git-svn https://svn.apache.org/repos/asf/httpd/httpd
    ghq get https://svn.apache.org/repos/asf/subversion
    ghq get --shallow hub.darcs.net/byorgey/split
    ghq get --bare x-motemen/gore
    ghq get --partial blobless x-motemen/blogsync
    ghq get --partial treeless x-motemen/gobump

    test -d $tmpdir/github.com/x-motemen/ghq/.git
    test -d $tmpdir/www.mercurial-scm.org/repo/hello/.hg
    test -d $tmpdir/launchpad.net/shutter/.bzr
    test -f $tmpdir/www.sqlite.org/src/.fslckout
    test -f $tmpdir/chiselapp.com/user/sti/repository/fossil-gui/.fslckout
    test -d $tmpdir/svn.apache.org/repos/asf/httpd/httpd/.git/svn
    test -d $tmpdir/svn.apache.org/repos/asf/subversion/.svn
    test -d $tmpdir/hub.darcs.net/byorgey/split/_darcs
    test -d $tmpdir/github.com/x-motemen/gore.git/refs
    grep --quiet "partialclonefilter = blob:none" $tmpdir/github.com/x-motemen/blogsync/.git/config
    grep --quiet "partialclonefilter = tree:0" $tmpdir/github.com/x-motemen/gobump/.git/config

: testing 'ghq list'
    cat <<EOF | sort > $tmpdir/expect
chiselapp.com/user/sti/repository/fossil-gui
github.com/x-motemen/blogsync
github.com/x-motemen/ghq
github.com/x-motemen/gobump
github.com/x-motemen/gore.git
www.mercurial-scm.org/repo/hello
launchpad.net/shutter
www.sqlite.org/src
svn.apache.org/repos/asf/httpd/httpd
svn.apache.org/repos/asf/subversion
hub.darcs.net/byorgey/split
EOF
    ghq list | sort > $tmpdir/got
    diff -u $tmpdir/expect $tmpdir/got

: testing 'input | ghq get -u'
    ghq list | ghq get -u

: testing 'ghq create'
    test "$(ghq create Songmu/hoge)" = "$tmpdir/github.com/Songmu/hoge"
    test -d $tmpdir/github.com/Songmu/hoge/.git
