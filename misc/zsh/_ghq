#compdef ghq

function _ghq () {
    local context curcontext=$curcontext state line
    declare -A opt_args
    local ret=1

    _arguments -C \
        '(-h --help)'{-h,--help}'[show help]' \
        '(-v --version)'{-v,--version}'[print the version]' \
        '1: :__ghq_commands' \
        '*:: :->args' \
        && ret=0

    case $state in
        (args)
            case $words[1] in
                (get)
                    _arguments -C \
                        '(-u --update)'{-u,--update}'[Update local repository if cloned already]' \
                        '-p[Clone with SSH]' \
                        '--shallow[Do a shallow clone]' \
                        '(-l --look)'{-l,--look}'[Look after get]' \
                        '--vcs[Specify vcs backend for cloning]' \
                        '(-s --silent)'{-s,--silent}'[Clone or update silently]' \
                        '--no-recursive[Prevent recursive fetching]' \
                        '(-b --branch)'{-b,--branch}'[Specify branch name]' \
                        '(-P --parallel)'{-P,--parallel}'[Import parallely]' \
                        '(-)*:: :->null_state' \
                        && ret=0
                    ;;
                (list)
                    _arguments -C \
                        '(-e --exact)'{-e,--exact}'[Perform an exact match]' \
                        '--vcs[Specify vcs backend for matching]' \
                        '(-p --full-path)'{-p,--full-path}'[Print full paths]' \
                        '--unique[Print unique subpaths]' \
                        '(-)*:: :->null_state' \
                        && ret=0
                    ;;
                (root)
                    _arguments -C \
                        '--all[Show all roots]' \
                        '(-)*:: :->null_state' \
                        && ret=0
                    ;;
                (create)
                    _arguments -C \
                        '--vcs[Specify vcs backend explicitly]' \
                        '(-)*:: :->null_state' \
                        && ret=0
                    ;;
                (help|h)
                    __ghq_commands && ret=0
                    ;;
            esac
            ;;
    esac

    return ret
}

__ghq_repositories () {
    local -a _repos
    _repos=( ${(@f)"$(_call_program repositories ghq list --unique)"} )
    _describe -t repositories Repositories _repos
}

__ghq_commands () {
    local -a _c
    _c=(
        'get:Clone/sync with a remote repository'
        'list:List local repositories'
        'create:Create a new repository'
        "root:Show repositories' root"
        'help:Show a list of commands or help for one command'
    )

    _describe -t commands Commands _c
}

_ghq "$@"
