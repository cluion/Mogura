package main

import (
	"errors"
	"fmt"

	"mogura/internal/i18n"
)

const bashCompletion = `# mogura bash completion
_mogura() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=($(compgen -W "clean analyze dev orphan monitor mem config completion version" -- "$cur"))
        return
    fi
    case "${COMP_WORDS[1]}" in
        clean|orphan) COMPREPLY=($(compgen -W "--list --json" -- "$cur")) ;;
        dev) COMPREPLY=($(compgen -W "--list --json" -- "$cur") $(compgen -d -- "$cur")) ;;
        analyze) COMPREPLY=($(compgen -d -- "$cur")) ;;
        mem) COMPREPLY=($(compgen -W "--json --drop-caches --swap-reset" -- "$cur")) ;;
        completion) COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur")) ;;
    esac
}
complete -F _mogura mogura
`

const zshCompletion = `#compdef mogura
_mogura() {
    local -a commands
    commands=(
        'clean:scan and clean system junk'
        'analyze:disk usage analyzer'
        'dev:scan build artifacts'
        'orphan:find orphaned configs'
        'monitor:live system monitor'
        'mem:top memory consumers'
        'config:open settings'
        'completion:print shell completion script'
        'version:show version'
    )
    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi
    case $words[2] in
        clean|orphan) _values 'flag' '--list' '--json' ;;
        dev) _alternative 'flags:flag:(--list --json)' 'dirs:directory:_files -/' ;;
        analyze) _files -/ ;;
        mem) _values 'flag' '--json' '--drop-caches' '--swap-reset' ;;
        completion) _values 'shell' bash zsh fish ;;
    esac
}
_mogura "$@"
`

const fishCompletion = `# mogura fish completion
complete -c mogura -f
complete -c mogura -n __fish_use_subcommand -a clean -d 'scan and clean system junk'
complete -c mogura -n __fish_use_subcommand -a analyze -d 'disk usage analyzer'
complete -c mogura -n __fish_use_subcommand -a dev -d 'scan build artifacts'
complete -c mogura -n __fish_use_subcommand -a orphan -d 'find orphaned configs'
complete -c mogura -n __fish_use_subcommand -a monitor -d 'live system monitor'
complete -c mogura -n __fish_use_subcommand -a mem -d 'top memory consumers'
complete -c mogura -n __fish_use_subcommand -a config -d 'open settings'
complete -c mogura -n __fish_use_subcommand -a completion -d 'print shell completion script'
complete -c mogura -n __fish_use_subcommand -a version -d 'show version'
complete -c mogura -n '__fish_seen_subcommand_from clean dev orphan' -l list
complete -c mogura -n '__fish_seen_subcommand_from clean dev orphan mem' -l json
complete -c mogura -n '__fish_seen_subcommand_from mem' -l drop-caches
complete -c mogura -n '__fish_seen_subcommand_from mem' -l swap-reset
complete -c mogura -n '__fish_seen_subcommand_from analyze dev' -F
complete -c mogura -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
`

func runCompletion(args []string) error {
	if len(args) != 1 {
		usage()
		return errors.New(i18n.T("completion 需要指定 shell:bash、zsh 或 fish"))
	}
	switch args[0] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		return fmt.Errorf(i18n.T("不支援的 shell: %s(支援 bash、zsh、fish)"), args[0])
	}
	return nil
}
