# Changelog

## [v1.6.1](https://github.com/x-motemen/ghq/compare/v1.6.0...v1.6.1) - 2024-04-10
- Revert "feat: enable relative path ssh URL" by @Songmu in https://github.com/x-motemen/ghq/pull/385

## [v1.6.0](https://github.com/x-motemen/ghq/compare/v1.5.0...v1.6.0) - 2024-04-04
- Feat: add alias for `ghq get` by @Okabe-Junya in https://github.com/x-motemen/ghq/pull/377
- Append ".git" to directory name for bare git repositories by @dochang in https://github.com/x-motemen/ghq/pull/368
- feat(misc): add fish completion by @folliehiyuki in https://github.com/x-motemen/ghq/pull/375
- Added mise-en-place installation and usage instructions. by @morihaya in https://github.com/x-motemen/ghq/pull/376
- ghq rm to support bare option by @Songmu in https://github.com/x-motemen/ghq/pull/381
- Go 1.22 and update deps by @Songmu in https://github.com/x-motemen/ghq/pull/382
- feat: enable relative path ssh URL by @JyJyJcr in https://github.com/x-motemen/ghq/pull/378

## [v1.5.0](https://github.com/x-motemen/ghq/compare/v1.4.2...v1.5.0) - 2024-02-02
- Support pijul by @yoichi in https://github.com/x-motemen/ghq/pull/372
- feat: implement `ghq rm` command by @Okabe-Junya in https://github.com/x-motemen/ghq/pull/371

## [v1.4.2](https://github.com/x-motemen/ghq/compare/v1.4.1...v1.4.2) - 2023-04-16
- release as draft by tagpr to hide it before uploading assets by @Songmu in https://github.com/x-motemen/ghq/pull/363
- de-capitalize errors to make staticcheck happy by @hezhizhen in https://github.com/x-motemen/ghq/pull/365
- adjustment by @hezhizhen in https://github.com/x-motemen/ghq/pull/366
- introduce codecov by @Songmu in https://github.com/x-motemen/ghq/pull/359
- update deps by @Songmu in https://github.com/x-motemen/ghq/pull/369

## [v1.4.1](https://github.com/x-motemen/ghq/compare/v1.4.0...v1.4.1) - 2023-02-22
- Add spell check workflow by @hezhizhen in https://github.com/x-motemen/ghq/pull/356
- Fix `samePath` by @hezhizhen in https://github.com/x-motemen/ghq/pull/357
- fix test on windows by @Songmu in https://github.com/x-motemen/ghq/pull/358
- update deps by @Songmu in https://github.com/x-motemen/ghq/pull/360
- Fixed a problem with SSH URLs, etc. by @Songmu in https://github.com/x-motemen/ghq/pull/361
- go-version-file: go.mod by @Songmu in https://github.com/x-motemen/ghq/pull/362

## [v1.4.0](https://github.com/x-motemen/ghq/compare/v1.3.0...v1.4.0) - 2023-02-22
- Use t.Cleanup to remove temporally directories and set environment variables by @itchyny in https://github.com/x-motemen/ghq/pull/343
- Support branch/tag via @ syntax by @NathanBaulch in https://github.com/x-motemen/ghq/pull/347
- introduce tagpr by @Songmu in https://github.com/x-motemen/ghq/pull/353

## [v1.3.0](https://github.com/x-motemen/ghq/compare/v1.2.1...v1.3.0) (2022-05-06)

* update deps [#342](https://github.com/x-motemen/ghq/pull/342) ([Songmu](https://github.com/Songmu))
* use go install instead of go get [#337](https://github.com/x-motemen/ghq/pull/337) ([shogo82148](https://github.com/shogo82148))
* Simplify repository lock using sync.Map [#339](https://github.com/x-motemen/ghq/pull/339) ([itchyny](https://github.com/itchyny))
* bump Go 1.18 [#338](https://github.com/x-motemen/ghq/pull/338) ([shogo82148](https://github.com/shogo82148))
* bump the actions that run on Node.js 16 runtime [#336](https://github.com/x-motemen/ghq/pull/336) ([shogo82148](https://github.com/shogo82148))
* fix: [trivial] fix the error message if nothing is passed to `ghq create` [#334](https://github.com/x-motemen/ghq/pull/334) ([gfx](https://github.com/gfx))
* feat: do a bare clone [#328](https://github.com/x-motemen/ghq/pull/328) ([helmecke](https://github.com/helmecke))
* update installation go [#327](https://github.com/x-motemen/ghq/pull/327) ([isanasan](https://github.com/isanasan))
* avoid using "Japanese Only" [#325](https://github.com/x-motemen/ghq/pull/325) ([shogo82148](https://github.com/shogo82148))

## [v1.2.1](https://github.com/x-motemen/ghq/compare/v1.2.0...v1.2.1) (2021-05-24)

* ~/.gitconfig incompatible with AWS CodeCommit region specific URL [#322](https://github.com/x-motemen/ghq/pull/322) ([xeres](https://github.com/xeres))

## [v1.2.0](https://github.com/x-motemen/ghq/compare/v1.1.7...v1.2.0) (2021-05-23)

* Support AWS CodeCommit HTTP (GRC) [#321](https://github.com/x-motemen/ghq/pull/321) ([xeres](https://github.com/xeres))
* Smartcasing [#317](https://github.com/x-motemen/ghq/pull/317) ([bmendric](https://github.com/bmendric))

## [v1.1.7](https://github.com/x-motemen/ghq/compare/v1.1.6...v1.1.7) (2021-02-27)

* sort list by default [#314](https://github.com/x-motemen/ghq/pull/314) ([Songmu](https://github.com/Songmu))

## [v1.1.6](https://github.com/x-motemen/ghq/compare/v1.1.5...v1.1.6) (2021-02-25)

* update deps [#313](https://github.com/x-motemen/ghq/pull/313) ([Songmu](https://github.com/Songmu))
* Add installation instructions for GNU Guix [#310](https://github.com/x-motemen/ghq/pull/310) ([lafrenierejm](https://github.com/lafrenierejm))
* bump saracen/walker v0.1.2 [#312](https://github.com/x-motemen/ghq/pull/312) ([shogo82148](https://github.com/shogo82148))
* Add installation using asdf-vm [#305](https://github.com/x-motemen/ghq/pull/305) ([kajisha](https://github.com/kajisha))
* Clone URLs by copying `net.URL` struct [#302](https://github.com/x-motemen/ghq/pull/302) ([vladimyr](https://github.com/vladimyr))
* Fix typo in `get` command log output [#301](https://github.com/x-motemen/ghq/pull/301) ([vladimyr](https://github.com/vladimyr))
* nit: Fix the pointer to example configuration [#298](https://github.com/x-motemen/ghq/pull/298) ([aereal](https://github.com/aereal))
* Add scoop at installation [#297](https://github.com/x-motemen/ghq/pull/297) ([hotchpotch](https://github.com/hotchpotch))

## [v1.1.5](https://github.com/x-motemen/ghq/compare/v1.1.4...v1.1.5) (2020-07-24)

* update Songmu/gitconfig [#295](https://github.com/x-motemen/ghq/pull/295) ([Songmu](https://github.com/Songmu))
* fix typo in readme [#293](https://github.com/x-motemen/ghq/pull/293) ([capytan](https://github.com/capytan))
* Add conda install instructions [#292](https://github.com/x-motemen/ghq/pull/292) ([sodre](https://github.com/sodre))

## [v1.1.4](https://github.com/x-motemen/ghq/compare/v1.1.3...v1.1.4) (2020-06-30)

* Care the some repository URL ended with '.git/' [#291](https://github.com/x-motemen/ghq/pull/291) ([yoichi](https://github.com/yoichi))
* Fix typo in `ghq get --help` [#290](https://github.com/x-motemen/ghq/pull/290) ([yoichi](https://github.com/yoichi))

## [v1.1.3](https://github.com/x-motemen/ghq/compare/v1.1.2...v1.1.3) (2020-06-26)

* Fix ghq root --all not showing all root directories [#289](https://github.com/x-motemen/ghq/pull/289) ([shihanng](https://github.com/shihanng))

## [v1.1.2](https://github.com/x-motemen/ghq/compare/v1.1.1...v1.1.2) (2020-06-21)

* update deps [#287](https://github.com/x-motemen/ghq/pull/287) ([Songmu](https://github.com/Songmu))
* Fetch instead of pull if the working tree isn't tracking remote. [#286](https://github.com/x-motemen/ghq/pull/286) ([yoichi](https://github.com/yoichi))
* Refactor so that RemoteRepository#VCS can return an error [#285](https://github.com/x-motemen/ghq/pull/285) ([johejo](https://github.com/johejo))

## [v1.1.1](https://github.com/x-motemen/ghq/compare/v1.1.0...v1.1.1) (2020-06-03)

* update deps [#284](https://github.com/x-motemen/ghq/pull/284) ([Songmu](https://github.com/Songmu))
* fix integration testing [#283](https://github.com/x-motemen/ghq/pull/283) ([Songmu](https://github.com/Songmu))
* Check vcsBackend is nil [#282](https://github.com/x-motemen/ghq/pull/282) ([mattn](https://github.com/mattn))
* Add installation note for Void Linux [#280](https://github.com/x-motemen/ghq/pull/280) ([imbsky](https://github.com/imbsky))
* Add bash completion to release archives [#279](https://github.com/x-motemen/ghq/pull/279) ([tottoto](https://github.com/tottoto))

## [v1.1.0](https://github.com/x-motemen/ghq/compare/v1.0.3...v1.1.0) (2020-01-24)

* update module name to github.com/x-motemen/ghq and fix import paths [#275](https://github.com/x-motemen/ghq/pull/275) ([Songmu](https://github.com/Songmu))

## [v1.0.3](https://github.com/motemen/ghq/compare/v1.0.2...v1.0.3) (2020-01-22)

* CGO_ENABLED=0 and drop -static flag [#273](https://github.com/motemen/ghq/pull/273) ([Songmu](https://github.com/Songmu))

## [v1.0.2](https://github.com/motemen/ghq/compare/v1.0.1...v1.0.2) (2020-01-17)

* static build [#271](https://github.com/motemen/ghq/pull/271) ([Songmu](https://github.com/Songmu))
* Add options to zsh-completion [#270](https://github.com/motemen/ghq/pull/270) ([darklore](https://github.com/darklore))
* update ghq.root description in README [#269](https://github.com/motemen/ghq/pull/269) ([ksoichiro](https://github.com/ksoichiro))
* Enhance vcs_test.go [#266](https://github.com/motemen/ghq/pull/266) ([Songmu](https://github.com/Songmu))
* fix testing for 'ghq get --look' [#265](https://github.com/motemen/ghq/pull/265) ([Songmu](https://github.com/Songmu))

## [v1.0.1](https://github.com/motemen/ghq/compare/v1.0.0...v1.0.1) (2020-01-05)

* Fix argument for look in get, to make `ghq get -look` work [#264](https://github.com/motemen/ghq/pull/264) ([astj](https://github.com/astj))

## [v1.0.0](https://github.com/motemen/ghq/compare/v0.99.2...v1.0.0) (2020-01-05)

* Release v1 [#262](https://github.com/motemen/ghq/pull/262) ([Songmu](https://github.com/Songmu))
* enhance testing around "ghq create" [#261](https://github.com/motemen/ghq/pull/261) ([Songmu](https://github.com/Songmu))
* start tests with empty gitconfig [#260](https://github.com/motemen/ghq/pull/260) ([Songmu](https://github.com/Songmu))

## [v0.99.2](https://github.com/motemen/ghq/compare/v0.99.1...v0.99.2) (2020-01-04)

* enhance logs on ghq get [#259](https://github.com/motemen/ghq/pull/259) ([Songmu](https://github.com/Songmu))
* fix semaphore [#258](https://github.com/motemen/ghq/pull/258) ([Songmu](https://github.com/Songmu))

## [v0.99.1](https://github.com/motemen/ghq/compare/v0.99.0...v0.99.1) (2020-01-02)

* detect VCS backend from URL scheme [#257](https://github.com/motemen/ghq/pull/257) ([Songmu](https://github.com/Songmu))
* print command execution output to stderr [#256](https://github.com/motemen/ghq/pull/256) ([Songmu](https://github.com/Songmu))

## [v0.99.0](https://github.com/motemen/ghq/compare/v0.17.4...v0.99.0) (2019-12-31)

* fix help docs [#255](https://github.com/motemen/ghq/pull/255) ([Songmu](https://github.com/Songmu))
* update documents [#252](https://github.com/motemen/ghq/pull/252) ([Songmu](https://github.com/Songmu))
* update deps [#251](https://github.com/motemen/ghq/pull/251) ([Songmu](https://github.com/Songmu))
* [incompatible] remove the "import" subcommand and merge this function into the "get" subcommand [#244](https://github.com/motemen/ghq/pull/244) ([Songmu](https://github.com/Songmu))
* [incompatible] drop look subcommand [#243](https://github.com/motemen/ghq/pull/243) ([Songmu](https://github.com/Songmu))
* [incompatible] make ~/ghq default root [#237](https://github.com/motemen/ghq/pull/237) ([Songmu](https://github.com/Songmu))
* add integration test for darcs [#250](https://github.com/motemen/ghq/pull/250) ([Songmu](https://github.com/Songmu))
* integration test on master [#249](https://github.com/motemen/ghq/pull/249) ([Songmu](https://github.com/Songmu))
* enhance svn and git-svn support [#248](https://github.com/motemen/ghq/pull/248) ([Songmu](https://github.com/Songmu))

## [v0.17.4](https://github.com/motemen/ghq/compare/v0.17.3...v0.17.4) (2019-12-29)

* [fix] uniq local roots [#246](https://github.com/motemen/ghq/pull/246) ([Songmu](https://github.com/Songmu))
* [bugfix] fix race conditions by using sync.Once [#245](https://github.com/motemen/ghq/pull/245) ([Songmu](https://github.com/Songmu))

## [v0.17.3](https://github.com/motemen/ghq/compare/v0.17.2...v0.17.3) (2019-12-26)

* [incompatible] fix localRepositoryRoots order [#240](https://github.com/motemen/ghq/pull/240) ([Songmu](https://github.com/Songmu))

## [v0.17.2](https://github.com/motemen/ghq/compare/v0.17.1...v0.17.2) (2019-12-26)

* adjust ghq.root detection [#238](https://github.com/motemen/ghq/pull/238) ([Songmu](https://github.com/Songmu))
* adjust document [#236](https://github.com/motemen/ghq/pull/236) ([Songmu](https://github.com/Songmu))
* stop providing 32bit binary [#235](https://github.com/motemen/ghq/pull/235) ([Songmu](https://github.com/Songmu))
* The parallel option is now an official feature [#234](https://github.com/motemen/ghq/pull/234) ([Songmu](https://github.com/Songmu))
* Update urfave/cli to v2 [#233](https://github.com/motemen/ghq/pull/233) ([Songmu](https://github.com/Songmu))

## [v0.17.1](https://github.com/motemen/ghq/compare/v0.17.0...v0.17.1) (2019-12-25)

* [bugfix] fix --no-recursive [#232](https://github.com/motemen/ghq/pull/232) ([Songmu](https://github.com/Songmu))
* Add bash-completion [#231](https://github.com/motemen/ghq/pull/231) ([mattn](https://github.com/mattn))

## [v0.17.0](https://github.com/motemen/ghq/compare/v0.16.0...v0.17.0) (2019-12-24)

* add create subcommand [#229](https://github.com/motemen/ghq/pull/229) ([Songmu](https://github.com/Songmu))
* ghq.<url>.root configuration [#228](https://github.com/motemen/ghq/pull/228) ([Songmu](https://github.com/Songmu))

## [v0.16.0](https://github.com/motemen/ghq/compare/v0.15.0...v0.16.0) (2019-12-22)

* fetch recursively by default on git repository and  add --no-recursive option to prevent it [#227](https://github.com/motemen/ghq/pull/227) ([moajo](https://github.com/moajo))
* add tests for package cmdutil [#226](https://github.com/motemen/ghq/pull/226) ([Songmu](https://github.com/Songmu))

## [v0.15.0](https://github.com/motemen/ghq/compare/v0.14.2...v0.15.0) (2019-12-19)

* refactor test helpers [#225](https://github.com/motemen/ghq/pull/225) ([Songmu](https://github.com/Songmu))
* drop deprecated `ghq.ghe.host` configuration [#224](https://github.com/motemen/ghq/pull/224) ([Songmu](https://github.com/Songmu))
* remove os.Exit from inside code [#223](https://github.com/motemen/ghq/pull/223) ([Songmu](https://github.com/Songmu))
* drop xerrors deps and update deps [#222](https://github.com/motemen/ghq/pull/222) ([Songmu](https://github.com/Songmu))
* Ignore and just log inaccessible directories instead of failing [#221](https://github.com/motemen/ghq/pull/221) ([Songmu](https://github.com/Songmu))
* adjust files structure [#220](https://github.com/motemen/ghq/pull/220) ([Songmu](https://github.com/Songmu))

## [v0.14.2](https://github.com/motemen/ghq/compare/v0.14.1...v0.14.2) (2019-12-18)

* treat git-svn as a kind of git repository in local [#219](https://github.com/motemen/ghq/pull/219) ([Songmu](https://github.com/Songmu))

## [v0.14.1](https://github.com/motemen/ghq/compare/v0.14.0...v0.14.1) (2019-12-18)

* update deps [#218](https://github.com/motemen/ghq/pull/218) ([Songmu](https://github.com/Songmu))
* refine local VCS detection order [#217](https://github.com/motemen/ghq/pull/217) ([Songmu](https://github.com/Songmu))
* [refactor] define type vcsGetOption struct [#216](https://github.com/motemen/ghq/pull/216) ([Songmu](https://github.com/Songmu))
* the import subcommand should not accept the --branch option [#214](https://github.com/motemen/ghq/pull/214) ([Songmu](https://github.com/Songmu))
* fix broken coverage report [#213](https://github.com/motemen/ghq/pull/213) ([shogo82148](https://github.com/shogo82148))
* Convertion for URL should return slash-ed paths. [#212](https://github.com/motemen/ghq/pull/212) ([mattn](https://github.com/mattn))

## [v0.14.0](https://github.com/motemen/ghq/compare/v0.13.1...v0.14.0) (2019-12-04)

* introduce GitHub Actions [#208](https://github.com/motemen/ghq/pull/208) ([Songmu](https://github.com/Songmu))
* Fix typos! [#210](https://github.com/motemen/ghq/pull/210) ([tockn](https://github.com/tockn))
* Fix redundant index [#209](https://github.com/motemen/ghq/pull/209) ([tockn](https://github.com/tockn))
* add Songmu to authors [#207](https://github.com/motemen/ghq/pull/207) ([Songmu](https://github.com/Songmu))
* Fix deprecation of urfave/cli [#206](https://github.com/motemen/ghq/pull/206) ([tockn](https://github.com/tockn))

## [v0.13.1](https://github.com/motemen/ghq/compare/v0.13.0...v0.13.1) (2019-12-04)

* optimize `--vcs` option of list [#205](https://github.com/motemen/ghq/pull/205) ([autopp](https://github.com/autopp))

## [v0.13.0](https://github.com/motemen/ghq/compare/v0.12.9...v0.13.0) (2019-12-01)

* update deps [#204](https://github.com/motemen/ghq/pull/204) ([Songmu](https://github.com/Songmu))
* Add --branch option to `ghq get` for specifying branch [#203](https://github.com/motemen/ghq/pull/203) ([rhysd](https://github.com/rhysd))

## [v0.12.9](https://github.com/motemen/ghq/compare/v0.12.8...v0.12.9) (2019-11-20)

* update deps [#201](https://github.com/motemen/ghq/pull/201) ([Songmu](https://github.com/Songmu))

## [v0.12.8](https://github.com/motemen/ghq/compare/v0.12.7...v0.12.8) (2019-11-05)

* [bugfix] lock in walking [#197](https://github.com/motemen/ghq/pull/197) ([Songmu](https://github.com/Songmu))

## [v0.12.7](https://github.com/motemen/ghq/compare/v0.12.6...v0.12.7) (2019-11-05)

* fix accepted values of ghq.<url>.vcs in README [#196](https://github.com/motemen/ghq/pull/196) ([autopp](https://github.com/autopp))
* Use walker [#195](https://github.com/motemen/ghq/pull/195) ([mattn](https://github.com/mattn))
* Use filepath [#194](https://github.com/motemen/ghq/pull/194) ([mattn](https://github.com/mattn))
* go.{mod,sum}: Update xerrors dependency [#193](https://github.com/motemen/ghq/pull/193) ([rvolosatovs](https://github.com/rvolosatovs))
* define func detectUserName and utilize it [#191](https://github.com/motemen/ghq/pull/191) ([Songmu](https://github.com/Songmu))

## [v0.12.6](https://github.com/motemen/ghq/compare/v0.12.5...v0.12.6) (2019-05-29)

* Removing port number from local directory. [#182](https://github.com/motemen/ghq/pull/182) ([gmidorii](https://github.com/gmidorii))

## [v0.12.5](https://github.com/motemen/ghq/compare/v0.12.4...v0.12.5) (2019-05-28)

* refine local cloning path detection [#181](https://github.com/motemen/ghq/pull/181) ([Songmu](https://github.com/Songmu))
* Adjust metafiles [#179](https://github.com/motemen/ghq/pull/179) ([Songmu](https://github.com/Songmu))

## [v0.12.4](https://github.com/motemen/ghq/compare/v0.12.3...v0.12.4) (2019-05-24)

* make archive [#178](https://github.com/motemen/ghq/pull/178) ([Songmu](https://github.com/Songmu))

## [v0.12.3](https://github.com/motemen/ghq/compare/v0.12.2...v0.12.3) (2019-05-16)

* [fix] Ignore files which seems to system hidden file in walking [#176](https://github.com/motemen/ghq/pull/176) ([Songmu](https://github.com/Songmu))
* [fix] Fix the list if symlink is in the same directory [#174](https://github.com/motemen/ghq/pull/174) ([at-grandpa](https://github.com/at-grandpa))
* [refactoring] introduce Songmu/gitconfig [#175](https://github.com/motemen/ghq/pull/175) ([Songmu](https://github.com/Songmu))
* [refactoring] Get ghq.completeUser strictly as a boolean value [#172](https://github.com/motemen/ghq/pull/172) ([Songmu](https://github.com/Songmu))

## [v0.12.2](https://github.com/motemen/ghq/compare/v0.12.1...v0.12.2) (2019-05-12)

* [fix] remove extra / from ref when path has "/" prefix [#167](https://github.com/motemen/ghq/pull/167) ([sakihet](https://github.com/sakihet))
* [fix] refine local cloning path detection [#171](https://github.com/motemen/ghq/pull/171) ([Songmu](https://github.com/Songmu))
* [fix] skip go-import mod in detectGoImport [#170](https://github.com/motemen/ghq/pull/170) ([Songmu](https://github.com/Songmu))
* [fix] even if the GHQ_ROOT directory doesn't exist, don't raise an error [#168](https://github.com/motemen/ghq/pull/168) ([Songmu](https://github.com/Songmu))

## [v0.12.1](https://github.com/motemen/ghq/compare/v0.12.0...v0.12.1) (2019-05-07)

* update deps [#164](https://github.com/motemen/ghq/pull/164) ([Songmu](https://github.com/Songmu))

## [v0.12.0](https://github.com/motemen/ghq/compare/v0.11.2...v0.12.0) (2019-05-07)

* get lock using repoPath to avoid duplicate get on import [#163](https://github.com/motemen/ghq/pull/163) ([Songmu](https://github.com/Songmu))
* separate gitutil package from main and utilize it [#162](https://github.com/motemen/ghq/pull/162) ([Songmu](https://github.com/Songmu))
* support deep directory in GitRepository like 'github.com/motemen/ghq/logger' [#161](https://github.com/motemen/ghq/pull/161) ([Songmu](https://github.com/Songmu))
* Test enhancement [#160](https://github.com/motemen/ghq/pull/160) ([Songmu](https://github.com/Songmu))
* add TestLocalRepository_VCS [#159](https://github.com/motemen/ghq/pull/159) ([Songmu](https://github.com/Songmu))

## [v0.11.2](https://github.com/motemen/ghq/compare/v0.11.1...v0.11.2) (2019-05-06)

* [bugfix] walk up directory when finding VCS of LocalRepository to avaid panics [#158](https://github.com/motemen/ghq/pull/158) ([Songmu](https://github.com/Songmu))
* bzr pull --overwrite same as go get [#157](https://github.com/motemen/ghq/pull/157) ([Songmu](https://github.com/Songmu))

## [v0.11.1](https://github.com/motemen/ghq/compare/v0.11.0...v0.11.1) (2019-05-05)

* [feature] List vcs option [#155](https://github.com/motemen/ghq/pull/155) ([msh5](https://github.com/msh5))
* [testing] add TestDoImport [#156](https://github.com/motemen/ghq/pull/156) ([Songmu](https://github.com/Songmu))
* [fix] fix findVCSBackend and add tests [#154](https://github.com/motemen/ghq/pull/154) ([Songmu](https://github.com/Songmu))
* [fix] Make sure the directory exists with the exact name [#145](https://github.com/motemen/ghq/pull/145) ([knu](https://github.com/knu))
* [bugfix] Fix fossil support [#153](https://github.com/motemen/ghq/pull/153) ([Songmu](https://github.com/Songmu))
* [testing] add TestDoList_query [#152](https://github.com/motemen/ghq/pull/152) ([Songmu](https://github.com/Songmu))
* [testing] add TestRunInDirSilently [#151](https://github.com/motemen/ghq/pull/151) ([Songmu](https://github.com/Songmu))
* [testing] add more tests in TestDoLook [#150](https://github.com/motemen/ghq/pull/150) ([Songmu](https://github.com/Songmu))
* [testing] add TestDoLook [#149](https://github.com/motemen/ghq/pull/149) ([Songmu](https://github.com/Songmu))
* [refactoring] remove NewFakeRunner which not used [#148](https://github.com/motemen/ghq/pull/148) ([Songmu](https://github.com/Songmu))
* [refactoring] Commonize doGet and doImport processing for refactoring [#147](https://github.com/motemen/ghq/pull/147) ([Songmu](https://github.com/Songmu))
* [testing] add test for `ghq root` [#146](https://github.com/motemen/ghq/pull/146) ([Songmu](https://github.com/Songmu))

## [v0.11.0](https://github.com/motemen/ghq/compare/v0.10.2...v0.11.0) (2019-05-01)

* Delete codes for the services turned down (Google Code and JazzHub) [#144](https://github.com/motemen/ghq/pull/144) ([Songmu](https://github.com/Songmu))
* reduce linter warnings [#143](https://github.com/motemen/ghq/pull/143) ([Songmu](https://github.com/Songmu))
* remove ghq.import.<subcommand> feature [#137](https://github.com/motemen/ghq/pull/137) ([Songmu](https://github.com/Songmu))
* Remove gomega [#142](https://github.com/motemen/ghq/pull/142) ([Songmu](https://github.com/Songmu))
* Streamline VCS detection for local repositories [#141](https://github.com/motemen/ghq/pull/141) ([Songmu](https://github.com/Songmu))
* introduce table driven test in TestCommandGet [#140](https://github.com/motemen/ghq/pull/140) ([Songmu](https://github.com/Songmu))
* "bzr" as Bazaar [#139](https://github.com/motemen/ghq/pull/139) ([Songmu](https://github.com/Songmu))
* Support bazaar [#87](https://github.com/motemen/ghq/pull/87) ([shigemk2](https://github.com/shigemk2))
* [experimental] support parallel import by using --parallel option [#136](https://github.com/motemen/ghq/pull/136) ([Songmu](https://github.com/Songmu))
* add -silent option to get and import [#135](https://github.com/motemen/ghq/pull/135) ([Songmu](https://github.com/Songmu))
* Stop exec on unix in the look subcommand in order to share code with windows [#134](https://github.com/motemen/ghq/pull/134) ([Songmu](https://github.com/Songmu))
* [list] accept url as a query [#133](https://github.com/motemen/ghq/pull/133) ([Songmu](https://github.com/Songmu))
* adjust NewURL() [#132](https://github.com/motemen/ghq/pull/132) ([Songmu](https://github.com/Songmu))

## [v0.10.2](https://github.com/motemen/ghq/compare/v0.10.1...v0.10.2) (2019-04-28)

* Fix `ghq root` [#131](https://github.com/motemen/ghq/pull/131) ([knu](https://github.com/knu))

## [v0.10.1](https://github.com/motemen/ghq/compare/v0.10.0...v0.10.1) (2019-04-28)

* Remove panics [#130](https://github.com/motemen/ghq/pull/130) ([Songmu](https://github.com/Songmu))
* generally loggers should output STDERR by default [#129](https://github.com/motemen/ghq/pull/129) ([Songmu](https://github.com/Songmu))
* Care the case if the query contains a hostname in list subcommand [#128](https://github.com/motemen/ghq/pull/128) ([Songmu](https://github.com/Songmu))
* don't panic when the current working directory doesn't exist. [#127](https://github.com/motemen/ghq/pull/127) ([Songmu](https://github.com/Songmu))
* canonicalize root paths to absolute [#126](https://github.com/motemen/ghq/pull/126) ([Songmu](https://github.com/Songmu))
* List symlinks to repos [#125](https://github.com/motemen/ghq/pull/125) ([Songmu](https://github.com/Songmu))
* rename package s/utils/cmdutil/ [#124](https://github.com/motemen/ghq/pull/124) ([Songmu](https://github.com/Songmu))
* separate logger package from utils [#123](https://github.com/motemen/ghq/pull/123) ([Songmu](https://github.com/Songmu))

## [v0.10.0](https://github.com/motemen/ghq/compare/v0.9.0...v0.10.0) (2019-04-27)

* drop mitchellh/go-homedir dependency [#122](https://github.com/motemen/ghq/pull/122) ([Songmu](https://github.com/Songmu))
* introduce Go Modules and adjust releng files [#121](https://github.com/motemen/ghq/pull/121) ([Songmu](https://github.com/Songmu))
* Add a dummy CVS backend to recognize and skip CVS working directories [#115](https://github.com/motemen/ghq/pull/115) ([knu](https://github.com/knu))
* add -l option on get command which immediately look after get [#112](https://github.com/motemen/ghq/pull/112) ([kuboon](https://github.com/kuboon))
* add support for Fossil SCM [#98](https://github.com/motemen/ghq/pull/98) ([motemen](https://github.com/motemen))
* Use parsed username also with ssh for Git [#101](https://github.com/motemen/ghq/pull/101) ([jjv](https://github.com/jjv))
* Add ghq.completeUser config to disable user completion of `ghq get` [#118](https://github.com/motemen/ghq/pull/118) ([k0kubun](https://github.com/k0kubun))
* ghq get --vcs=<vcs> [#72](https://github.com/motemen/ghq/pull/72) ([motemen](https://github.com/motemen))
* warn if executable was not found when RunCommand [#70](https://github.com/motemen/ghq/pull/70) ([motemen](https://github.com/motemen))
* support `meta name="go-import"` to detect Go repository [#120](https://github.com/motemen/ghq/pull/120) ([Songmu](https://github.com/Songmu))
* support refs which start with URL Authority in ghq get [#119](https://github.com/motemen/ghq/pull/119) ([Songmu](https://github.com/Songmu))

## [v0.9.0](https://github.com/motemen/ghq/compare/v0.8.0...v0.9.0) (2018-11-26)

* Use new constructor for logger [#104](https://github.com/motemen/ghq/pull/104) ([raviqqe](https://github.com/raviqqe))
* fix typo direcotry -> directory [#93](https://github.com/motemen/ghq/pull/93) ([naofumi-fujii](https://github.com/naofumi-fujii))

## [v0.8.0](https://github.com/motemen/ghq/compare/v0.7.4...v0.8.0) (2017-08-22)

- [breaking feature] If given URL does not contain / character, treat the URL as `https://github.com/<USERNAME>/<URL>`, where USERNAME is GitHub username obtained from `ghq.user` Git configuration variable, GITHUB_USER or USER (USERNAME in Windows) environment variables thanks to @b4b4r07 (#81)
- [maintenance] Fix building configuration thanks to @south37 (#85), @smizy (#82)

## [v0.7.4](https://github.com/motemen/ghq/compare/v0.7.3...v0.7.4) (2016-03-07)

* support path list in GHQ_ROOT [#71](https://github.com/motemen/ghq/pull/71) ([hatotaka](https://github.com/hatotaka))

## [v0.7.3](https://github.com/motemen/ghq/compare/v0.7.2...v0.7.3) (2016-03-02)

* Github relative [#43](https://github.com/motemen/ghq/pull/43) ([mattn](https://github.com/mattn))

## [v0.7.2](https://github.com/motemen/ghq/compare/v0.7.1...v0.7.2) (2015-12-11)

* Revert "Merge pull request #54 from maoe/skip-non-vcs-dirs" [#66](https://github.com/motemen/ghq/pull/66) ([motemen](https://github.com/motemen))

## [v0.7.1](https://github.com/motemen/ghq/compare/v0.7...v0.7.1) (2015-08-06)

* Fix an issue of listing with directories containing symlinks [#61](https://github.com/motemen/ghq/pull/61) ([motemen](https://github.com/motemen))

## [v0.7](https://github.com/motemen/ghq/compare/v0.6...v0.7) (2015-08-03)

* Support for Bluemix DevOps Git service [#56](https://github.com/motemen/ghq/pull/56) ([uetchy](https://github.com/uetchy))
* GHQ_ROOT environment variable to override the root [#59](https://github.com/motemen/ghq/pull/59) ([motemen](https://github.com/motemen))
* Add darcs backend [#55](https://github.com/motemen/ghq/pull/55) ([maoe](https://github.com/maoe))
* fix failing test [#58](https://github.com/motemen/ghq/pull/58) ([motemen](https://github.com/motemen))
* Skip non-VCS directories for performance [#54](https://github.com/motemen/ghq/pull/54) ([maoe](https://github.com/maoe))
* fix test [#57](https://github.com/motemen/ghq/pull/57) ([motemen](https://github.com/motemen))
* `look` command accepts remote repository url too. [#51](https://github.com/motemen/ghq/pull/51) ([ryotarai](https://github.com/ryotarai))
* Add GHQ_LOOK env variable to a new shell executed by `ghq look` [#47](https://github.com/motemen/ghq/pull/47) ([superbrothers](https://github.com/superbrothers))

## [v0.6](https://github.com/motemen/ghq/compare/v0.5...v0.6) (2014-11-20)

* support gist URLs [#46](https://github.com/motemen/ghq/pull/46) ([motemen](https://github.com/motemen))
* Return exit status 1 for clone failure [#45](https://github.com/motemen/ghq/pull/45) ([k0kubun](https://github.com/k0kubun))

## [v0.5](https://github.com/motemen/ghq/compare/v0.4...v0.5) (2014-10-11)

* fixup docs and zsh completion [#44](https://github.com/motemen/ghq/pull/44) ([motemen](https://github.com/motemen))
* Add 'root' subcommand completion [#42](https://github.com/motemen/ghq/pull/42) ([syohex](https://github.com/syohex))
* Include zsh completion into release zip files [#41](https://github.com/motemen/ghq/pull/41) ([itiut](https://github.com/itiut))
* Add --all option to the root command [#40](https://github.com/motemen/ghq/pull/40) ([aaa707](https://github.com/aaa707))
* import: Accept the same clone flags with get command [#37](https://github.com/motemen/ghq/pull/37) ([eagletmt](https://github.com/eagletmt))
* accept SCP-like URL (git@github.com) for import command [#35](https://github.com/motemen/ghq/pull/35) ([mkanai](https://github.com/mkanai))
* Add root command [#34](https://github.com/motemen/ghq/pull/34) ([aaa707](https://github.com/aaa707))
* Set exit code of `look` which failed [#33](https://github.com/motemen/ghq/pull/33) ([fujimura](https://github.com/fujimura))
* Re-implement `ghq import` [#31](https://github.com/motemen/ghq/pull/31) ([motemen](https://github.com/motemen))
* use go-homedir for distributing compiled binaries [#32](https://github.com/motemen/ghq/pull/32) ([motemen](https://github.com/motemen))
* Fix for latest github.com/codegangsta/cli [#28](https://github.com/motemen/ghq/pull/28) ([syohex](https://github.com/syohex))

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
