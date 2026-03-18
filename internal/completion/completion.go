package completion

import "fmt"

// Generate returns the shell completion script for the given shell.
// Supported shells: bash, zsh, fish.
func Generate(shell string) (string, error) {
	switch shell {
	case "bash":
		return bash, nil
	case "zsh":
		return zsh, nil
	case "fish":
		return fish, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (use bash, zsh, or fish)", shell)
	}
}

const bash = `_files2clip() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--version --verbose --full-paths --include-binary --from-clipboard --file -f --exclude -e --ignore-file --max-file-size --max-total-size --max-files --completion"

    case "$prev" in
        --file|-f|--ignore-file)
            COMPREPLY=( $(compgen -f -- "$cur") )
            return 0
            ;;
        --max-file-size|--max-total-size)
            COMPREPLY=( $(compgen -W "1MB 5MB 10MB 50MB 100MB 1GB" -- "$cur") )
            return 0
            ;;
        --max-files)
            COMPREPLY=( $(compgen -W "100 500 1000 5000" -- "$cur") )
            return 0
            ;;
        --completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            return 0
            ;;
        --exclude|-e)
            return 0
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
    else
        COMPREPLY=( $(compgen -f -- "$cur") )
    fi
}
complete -o default -F _files2clip files2clip
`

const zsh = `#compdef files2clip

_files2clip() {
    _arguments \
        '--version[print version and exit]' \
        '--verbose[show detailed processing info]' \
        '--full-paths[use absolute paths in output]' \
        '--include-binary[include binary files]' \
        '--from-clipboard[read paths from clipboard]' \
        '(-f --file)'{-f,--file}'[read paths from a file]:file:_files' \
        '*'{-e,--exclude}'[exclude pattern]:pattern:' \
        '--ignore-file[gitignore-style exclusion file]:file:_files' \
        '--max-file-size[max individual file size]:size:(1MB 5MB 10MB 50MB 100MB 1GB)' \
        '--max-total-size[max total content size]:size:(1MB 5MB 10MB 50MB 100MB 1GB)' \
        '--max-files[max number of files]:count:(100 500 1000 5000)' \
        '--completion[generate shell completion]:shell:(bash zsh fish)' \
        '*:path:_files'
}

_files2clip "$@"
`

const fish = `complete -c files2clip -l version -d 'Print version and exit'
complete -c files2clip -l verbose -d 'Show detailed processing info'
complete -c files2clip -l full-paths -d 'Use absolute paths in output'
complete -c files2clip -l include-binary -d 'Include binary files'
complete -c files2clip -l from-clipboard -d 'Read paths from clipboard'
complete -c files2clip -s f -l file -r -F -d 'Read paths from a file'
complete -c files2clip -s e -l exclude -r -d 'Exclude pattern'
complete -c files2clip -l ignore-file -r -F -d 'Gitignore-style exclusion file'
complete -c files2clip -l max-file-size -r -d 'Max individual file size'
complete -c files2clip -l max-total-size -r -d 'Max total content size'
complete -c files2clip -l max-files -r -d 'Max number of files'
complete -c files2clip -l completion -r -f -a 'bash zsh fish' -d 'Generate shell completion'
`
