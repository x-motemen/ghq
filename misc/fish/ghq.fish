function __fish_ghq_needs_subcommand
    set -l cmd (commandline -opc)
    for subcmd in get list rm root create h help
        if contains -- $subcmd $cmd
            return 1
        end
    end
    return 0
end

# Remove any previous completion
complete -c ghq -e
# Don't suggest files
complete -c ghq -f

# Global arguments
complete -c ghq -s h -l help -d 'Show help'
complete -c ghq -n __fish_ghq_needs_subcommand -s v -l version -d 'Print the version'

# Global subcommands
complete -c ghq -n __fish_ghq_needs_subcommand -a get -d 'Clone/sync with a remote repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a list -d 'List local repositories'
complete -c ghq -n __fish_ghq_needs_subcommand -a rm -d 'Remove local repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a root -d 'Show repositories\' root'
complete -c ghq -n __fish_ghq_needs_subcommand -a create -d 'Create a new repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a 'h help' -d 'Shows a list of commands or help for one command'

# Arguments for subcommands
complete -c ghq -n '__fish_seen_subcommand_from get' -s u -l update -d 'Update local repository if cloned already'
complete -c ghq -n '__fish_seen_subcommand_from get' -s p -d 'Clone with SSH'
complete -c ghq -n '__fish_seen_subcommand_from get' -l shallow -d 'Do a shallow clone'
complete -c ghq -n '__fish_seen_subcommand_from get' -s l -l look -d 'Look after get'
complete -c ghq -n '__fish_seen_subcommand_from get' -l vcs -d 'Specify vcs backend for cloning'
complete -c ghq -n '__fish_seen_subcommand_from get' -s s -l silent -d 'Clone or update silently'
complete -c ghq -n '__fish_seen_subcommand_from get' -l no-recursive -d 'Prevent recursive fetching'
complete -c ghq -n '__fish_seen_subcommand_from get' -s b -l branch -d 'Specify branch name. This flag implies --single-branch on Git'
complete -c ghq -n '__fish_seen_subcommand_from get' -s P -l parallel -d 'Import parallelly'
complete -c ghq -n '__fish_seen_subcommand_from get' -l bare -d 'Do a bare clone'

complete -c ghq -n '__fish_seen_subcommand_from list' -s e -l exact -d 'Perform an exact match'
complete -c ghq -n '__fish_seen_subcommand_from list' -l vcs -d 'Specify vcs backend for matching'
complete -c ghq -n '__fish_seen_subcommand_from list' -s p -l full-path -d 'Print full paths'
complete -c ghq -n '__fish_seen_subcommand_from list' -l unique -d 'Print unique subpaths'

complete -c ghq -n '__fish_seen_subcommand_from rm' -l dry-run -d 'Do not remove actually'

complete -c ghq -n '__fish_seen_subcommand_from root' -l all -d 'Show all roots'

complete -c ghq -n '__fish_seen_subcommand_from create' -l vcs -d 'Specify vcs backend explicitly'

# Complete VCS backend options for supported subcommands
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'git github codecommit' -d git
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'svn subversion' -d subversion
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a git-svn -d git-svn
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'hg mercurial' -d mercurial
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a darcs -d darcs
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a pijul -d pijul
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a fossil -d fossil
complete -c ghq -n '__fish_seen_subcommand_from get list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'bzr bazaar' -d bazaar
