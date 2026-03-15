package cmd

import (
	"fmt"
	"os"
)

// RunCompletion handles the "completion" subcommand.
func RunCompletion(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: peerclaw completion <bash|zsh|fish>")
		return 1
	}

	switch args[0] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s (use bash, zsh, or fish)\n", args[0])
		return 1
	}
	return 0
}

const bashCompletion = `# peerclaw bash completion
# Add to ~/.bashrc: eval "$(peerclaw completion bash)"
_peerclaw() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="agent invoke inbox send health config reputation identity send-file transfer mcp acp notifications completion version help"
    agent_sub="list get register claim delete update discover heartbeat verify"
    inbox_sub="request status list"
    config_sub="set get list"
    reputation_sub="show list"
    identity_sub="anchor verify"
    notifications_sub="list count read read-all"
    transfer_sub="status"

    case "${prev}" in
        peerclaw)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return 0
            ;;
        agent)
            COMPREPLY=($(compgen -W "${agent_sub}" -- "${cur}"))
            return 0
            ;;
        inbox)
            COMPREPLY=($(compgen -W "${inbox_sub}" -- "${cur}"))
            return 0
            ;;
        config)
            COMPREPLY=($(compgen -W "${config_sub}" -- "${cur}"))
            return 0
            ;;
        reputation)
            COMPREPLY=($(compgen -W "${reputation_sub}" -- "${cur}"))
            return 0
            ;;
        identity)
            COMPREPLY=($(compgen -W "${identity_sub}" -- "${cur}"))
            return 0
            ;;
        notifications)
            COMPREPLY=($(compgen -W "${notifications_sub}" -- "${cur}"))
            return 0
            ;;
        transfer)
            COMPREPLY=($(compgen -W "${transfer_sub}" -- "${cur}"))
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return 0
            ;;
    esac

    # Global flags for most subcommands.
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--server --token --output" -- "${cur}"))
        return 0
    fi
}
complete -F _peerclaw peerclaw
`

const zshCompletion = `#compdef peerclaw
# peerclaw zsh completion
# Add to ~/.zshrc: eval "$(peerclaw completion zsh)"

_peerclaw() {
    local -a commands agent_sub inbox_sub config_sub reputation_sub identity_sub notifications_sub transfer_sub

    commands=(
        'agent:Manage agents'
        'invoke:Invoke an agent'
        'inbox:Manage access requests'
        'send:Send a message through the bridge'
        'health:Check server health'
        'config:Manage CLI configuration'
        'reputation:Reputation scores'
        'identity:Identity anchoring'
        'send-file:Send a file to another agent'
        'transfer:Manage file transfers'
        'mcp:MCP server for AI tool integration'
        'acp:ACP stdio bridge'
        'notifications:Manage notifications'
        'completion:Generate shell completion'
        'version:Print version'
        'help:Show help'
    )

    agent_sub=(list get register claim delete update discover heartbeat verify)
    inbox_sub=(request status list)
    config_sub=(set get list)
    reputation_sub=(show list)
    identity_sub=(anchor verify)
    notifications_sub=(list count read read-all)
    transfer_sub=(status)

    _arguments -C \
        '1:command:->command' \
        '*::arg:->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case ${words[1]} in
                agent) _describe 'subcommand' agent_sub ;;
                inbox) _describe 'subcommand' inbox_sub ;;
                config) _describe 'subcommand' config_sub ;;
                reputation) _describe 'subcommand' reputation_sub ;;
                identity) _describe 'subcommand' identity_sub ;;
                notifications) _describe 'subcommand' notifications_sub ;;
                transfer) _describe 'subcommand' transfer_sub ;;
                completion) _describe 'shell' '(bash zsh fish)' ;;
            esac
            ;;
    esac
}

compdef _peerclaw peerclaw
`

const fishCompletion = `# peerclaw fish completion
# Add to ~/.config/fish/completions/peerclaw.fish
# Or run: peerclaw completion fish > ~/.config/fish/completions/peerclaw.fish

# Disable file completions by default.
complete -c peerclaw -f

# Top-level commands.
complete -c peerclaw -n '__fish_use_subcommand' -a 'agent' -d 'Manage agents'
complete -c peerclaw -n '__fish_use_subcommand' -a 'invoke' -d 'Invoke an agent'
complete -c peerclaw -n '__fish_use_subcommand' -a 'inbox' -d 'Manage access requests'
complete -c peerclaw -n '__fish_use_subcommand' -a 'send' -d 'Send a message through the bridge'
complete -c peerclaw -n '__fish_use_subcommand' -a 'health' -d 'Check server health'
complete -c peerclaw -n '__fish_use_subcommand' -a 'config' -d 'Manage CLI configuration'
complete -c peerclaw -n '__fish_use_subcommand' -a 'reputation' -d 'Reputation scores'
complete -c peerclaw -n '__fish_use_subcommand' -a 'identity' -d 'Identity anchoring'
complete -c peerclaw -n '__fish_use_subcommand' -a 'send-file' -d 'Send a file to another agent'
complete -c peerclaw -n '__fish_use_subcommand' -a 'transfer' -d 'Manage file transfers'
complete -c peerclaw -n '__fish_use_subcommand' -a 'mcp' -d 'MCP server for AI tool integration'
complete -c peerclaw -n '__fish_use_subcommand' -a 'acp' -d 'ACP stdio bridge'
complete -c peerclaw -n '__fish_use_subcommand' -a 'notifications' -d 'Manage notifications'
complete -c peerclaw -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completion'
complete -c peerclaw -n '__fish_use_subcommand' -a 'version' -d 'Print version'
complete -c peerclaw -n '__fish_use_subcommand' -a 'help' -d 'Show help'

# agent subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from agent' -a 'list get register claim delete update discover heartbeat verify'

# inbox subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from inbox' -a 'request status list'

# config subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from config' -a 'set get list'

# reputation subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from reputation' -a 'show list'

# identity subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from identity' -a 'anchor verify'

# notifications subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from notifications' -a 'list count read read-all'

# transfer subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from transfer' -a 'status'

# completion subcommands.
complete -c peerclaw -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# Global flags.
complete -c peerclaw -l server -d 'PeerClaw server URL'
complete -c peerclaw -l token -d 'JWT auth token'
complete -c peerclaw -l output -d 'Output format (table or json)'
`
