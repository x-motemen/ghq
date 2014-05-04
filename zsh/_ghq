#compdef ghq

function _ghq () {
    _arguments -C \
        '1:: :__ghq_commands' \
        '2:: :->args' \
        && return 0

    case $state in
        (args)
            case $line[1] in
                (look)
                    _repos=( ${(f)"$(ghq list --unique)":gs/:/\\:/} )
                    _describe Repositories _repos
                    ;;
            esac
    esac

    return 1
}

__ghq_commands () {
    _c=(
        ${(f)"$(ghq help | perl -nle 'print qq(${1}[$2]) if /^COMMANDS:/.../^\S/ and /^   (\w+)(?:, \w+)?\t+(.+)/')"}
    )

    _values Commands $_c
}

compdef _ghq ghq
