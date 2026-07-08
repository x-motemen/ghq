function __fish_ghq_needs_subcommand
    set -l cmd (commandline -opc)
    for subcmd in get clone list rm root create migrate h help
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
complete -c ghq -n __fish_ghq_needs_subcommand -a 'get clone' -d 'Clone/sync with a remote repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a list -d 'List local repositories'
complete -c ghq -n __fish_ghq_needs_subcommand -a rm -d 'Remove local repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a root -d 'Show repositories\' root'
complete -c ghq -n __fish_ghq_needs_subcommand -a create -d 'Create a new repository'
complete -c ghq -n __fish_ghq_needs_subcommand -a migrate -d 'Migrate existing repository to ghq-managed directory'
complete -c ghq -n __fish_ghq_needs_subcommand -a 'h help' -d 'Shows a list of commands or help for one command'

# Arguments for subcommands
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s u -l update -d 'Update local repository if cloned already'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s p -d 'Clone with SSH'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -l shallow -d 'Do a shallow clone'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s l -l look -d 'Look after get'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -l vcs -d 'Specify vcs backend for cloning'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s s -l silent -d 'Clone or update silently'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -l no-recursive -d 'Prevent recursive fetching'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s b -l branch -d 'Specify branch name. This flag implies --single-branch on Git'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -s P -l parallel -d 'Import parallelly'
complete -c ghq -n '__fish_seen_subcommand_from get clone' -l bare -d 'Do a bare clone'
function __complete_get_partial
    printf '%s\t%s\n' 'blobless' 'Do a blobless clone'
    printf '%s\t%s\n' 'treeless' 'Do a treeless clone'
end
complete -c ghq -n '__fish_seen_subcommand_from get clone' -l partial -d 'Do a partial clone' -xa '(__complete_get_partial)'
# When updating an existing repository (-u/--update), complete with local repositories
complete -c ghq -n '__fish_seen_subcommand_from get clone' -n '__fish_seen_argument -s u -l update' -xa '(ghq list)'

complete -c ghq -n '__fish_seen_subcommand_from list' -s e -l exact -d 'Perform an exact match'
complete -c ghq -n '__fish_seen_subcommand_from list' -l vcs -d 'Specify vcs backend for matching'
complete -c ghq -n '__fish_seen_subcommand_from list' -s p -l full-path -d 'Print full paths'
complete -c ghq -n '__fish_seen_subcommand_from list' -l unique -d 'Print unique subpaths'
complete -c ghq -n '__fish_seen_subcommand_from list' -l bare -d 'Query bare repositories'

complete -c ghq -n '__fish_seen_subcommand_from rm' -l dry-run -d 'Do not remove actually'
complete -c ghq -n '__fish_seen_subcommand_from rm' -l bare -d 'Remove a bare repository'
complete -c ghq -n '__fish_seen_subcommand_from rm' -xa '(ghq list)'

complete -c ghq -n '__fish_seen_subcommand_from root' -l all -d 'Show all roots'

complete -c ghq -n '__fish_seen_subcommand_from create' -l vcs -d 'Specify vcs backend explicitly'
complete -c ghq -n '__fish_seen_subcommand_from create' -l bare -d 'Create a bare repository'

complete -c ghq -n '__fish_seen_subcommand_from migrate' -s y -d 'Skip confirmation prompt'
complete -c ghq -n '__fish_seen_subcommand_from migrate' -l dry-run -d 'Show what would happen without moving'

# Complete VCS backend options for supported subcommands
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'git github codecommit' -d git
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'svn subversion' -d subversion
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a git-svn -d git-svn
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'hg mercurial' -d mercurial
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a darcs -d darcs
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a pijul -d pijul
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a fossil -d fossil
complete -c ghq -n '__fish_seen_subcommand_from get clone list create' -n '__fish_seen_argument --vcs' -l vcs -x -a 'bzr bazaar' -d bazaar
