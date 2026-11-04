#!/bin/sh
set -e

tmpdir=$(mktemp -d)
tmpdir=$(cd "$tmpdir" && pwd -P)

cleanup() {
    code=$?
    rm -rf "$tmpdir"
    exit $code
}
trap cleanup EXIT

set -x

export GHQ_ROOT="$tmpdir"

: testing 'ghq get'
    ghq get x-motemen/ghq
    ghq get --vcs hg https://foss.heptapod.net/mercurial/hg-git
    ghq get https://launchpad.net/shutter
    ghq get --vcs fossil https://sqlite.org/althttpd
    ghq get --shallow --vcs=git-svn https://svn.apache.org/repos/asf/serf/trunk
    ghq get https://svn.apache.org/repos/asf/apr/apr
    ghq get --shallow --vcs darcs https://smeder.ee/~toastal/nixtamal.darcs
    ghq get --bare x-motemen/gore
    ghq get --partial blobless x-motemen/blogsync
    ghq get --partial treeless x-motemen/gobump

    test -d "$tmpdir/github.com/x-motemen/ghq/.git"
    test -d "$tmpdir/foss.heptapod.net/mercurial/hg-git/.hg"
    test -d "$tmpdir/launchpad.net/shutter/.bzr"
    test -f "$tmpdir/sqlite.org/althttpd/.fslckout"
    test -d "$tmpdir/svn.apache.org/repos/asf/serf/.git/svn"
    test -d "$tmpdir/svn.apache.org/repos/asf/apr/apr/.svn"
    test -d "$tmpdir/smeder.ee/~toastal/nixtamal.darcs/_darcs"
    test -d "$tmpdir/github.com/x-motemen/gore.git/refs"
    grep --quiet "partialclonefilter = blob:none" "$tmpdir/github.com/x-motemen/blogsync/.git/config"
    grep --quiet "partialclonefilter = tree:0" "$tmpdir/github.com/x-motemen/gobump/.git/config"

: testing 'ghq list'
    cat <<EOF | sort > "$tmpdir/expect"
github.com/x-motemen/blogsync
github.com/x-motemen/ghq
github.com/x-motemen/gobump
github.com/x-motemen/gore.git
foss.heptapod.net/mercurial/hg-git
launchpad.net/shutter
sqlite.org/althttpd
svn.apache.org/repos/asf/serf
svn.apache.org/repos/asf/apr/apr
smeder.ee/~toastal/nixtamal.darcs
EOF
    ghq list | sort > "$tmpdir/got"
    diff -u "$tmpdir/expect" "$tmpdir/got"

: testing 'input | ghq get -u'
    ghq list | ghq get -u

: testing 'ghq create'
    test "$(ghq create Songmu/hoge)" = "$tmpdir/github.com/Songmu/hoge"
    test -d "$tmpdir/github.com/Songmu/hoge/.git"
