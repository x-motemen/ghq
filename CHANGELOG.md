# Changelog

## 0.8.0 (2017-08-22)

- [breaking feature] If given URL does not contain / character, treat the URL as `https://github.com/<USERNAME>/<URL>`, where USERNAME is GitHub username obtained from `ghq.user` Git configuration variable, GITHUB_USER or USER (USERNAME in Windows) environment variables thanks to @b4b4r07 (#81)
- [maintenance] Fix building configuration thanks to @south37 (#85), @smizy (#82)

## 0.7

<History lost>

## 0.4 (2014-06-26)

- [feature] Support per-URL configuration variables e.g. `ghq.<URL>.vcs` to skip VCS backend auto-detection
- [fix] Fixed path problems of SCP-like URLs thanks to @osamu2001 (#20)
- [fix] `ghq get -u` now updates work tree for Mercurial repositories thanks to @troter (#19)
- And typo fixes thanks to @sorah, @dtan4 (#17, #18)

## 0.3 (2014-06-17)

- [feature] `ghq get -shallow` to perform a shallow clone
- [feature] Use GitHub token for `ghq import starred` if specified thanks to @makimoto (#16)
- [fix] Resolve ghq.root's symlinks thanks to @sorah (#15)

## 0.2 (2014-06-10)

- [feature] Support SCP-like repository URLs thanks to @kentaro (#1)
- [feature] Support GitHub:Enterprise repository URLs thanks to @kentaro (#2)
- [fix] Fix issue that default config variable was never used thanks to @Sixeight (#3)
- [fix] Support Windows environment thanks to @mattn (#5)
- [feature] `ghq get -p` to clone GitHub repositories with SSH thanks to @moznion (#7)
- [feature] Support any remotes other than GitHub and Google Code thanks to @tcnksm (#8, #13)
- [feature] Improve zsh completion thanks to @mollifier (#12)
- [feature] Support `ghq get git` for GitHub repositories with user and project name same thanks to @Sixeight (#14)
- And documentation updates thanks to @kentaro, @tricknotes (#6, #9)

## 0.1 (2014-06-01)

- Initial release
