package cmd

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func addTokenFlag(fs *flag.FlagSet, token *string) {
	t := os.Getenv("PEERCLAW_TOKEN")
	fs.StringVar(token, "token", t, "JWT auth token (or PEERCLAW_TOKEN env)")
}

// Version is set by GoReleaser via ldflags at build time.
var Version = "dev"

const defaultServer = "http://localhost:8080"

// Run executes the CLI with the given arguments.
func Run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 1
	}

	// Server URL priority: env var > config file > default.
	serverURL := os.Getenv("PEERCLAW_SERVER")
	if serverURL == "" {
		if cfg, err := loadCLIConfig(); err == nil && cfg.Server != "" {
			serverURL = cfg.Server
		} else {
			serverURL = defaultServer
		}
	}

	if err := validateServerURL(serverURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	switch args[0] {
	case "agent":
		return RunAgent(args[1:], serverURL)
	case "invoke":
		return RunInvoke(args[1:], serverURL)
	case "inbox":
		return RunInbox(args[1:], serverURL)
	case "send":
		return RunSend(args[1:], serverURL)
	case "health":
		return RunHealth(args[1:], serverURL)
	case "config":
		return RunConfig(args[1:])
	case "reputation":
		return RunReputation(args[1:], serverURL)
	case "identity":
		return RunIdentity(args[1:], serverURL)
	case "send-file":
		return RunSendFile(args[1:], serverURL)
	case "transfer":
		return RunTransfer(args[1:], serverURL)
	case "mcp":
		return RunMCP(args[1:], serverURL)
	case "acp":
		return RunACP(args[1:], serverURL)
	case "notifications":
		return RunNotifications(args[1:], serverURL)
	case "service":
		return RunService(args[1:], serverURL)
	case "completion":
		return RunCompletion(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version":
		fmt.Printf("peerclaw version %s\n", Version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw <command> [options]

Commands:
  agent       Manage agents (list, get, register, update, claim, discover, heartbeat, verify)
  invoke      Invoke an agent (send message and get response)
  inbox       Manage access requests (request, status, list)
  send        Send a message through the bridge
  health      Check server health
  config      Manage CLI configuration
  reputation  Reputation scores (show, list)
  send-file   Send a file to another agent (P2P)
  transfer    Manage file transfers (status)
  identity    Identity anchoring (anchor, verify)
  mcp         MCP server for AI tool integration (serve)
  acp         ACP stdio bridge for agent communication (serve)
  notifications Manage notifications (list, count, read, read-all)
  service     Manage OS-level agent service (install, uninstall, status, logs)
  completion  Generate shell completion (bash, zsh, fish)
  version     Print version

Environment:
  PEERCLAW_SERVER   Server URL (default: %s)
  PEERCLAW_TOKEN    JWT auth token for authenticated commands

Use "peerclaw <command> -h" for more information.
`, defaultServer)
}

func validateServerURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("server URL must use http or https scheme, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("server URL must have a host")
	}
	return nil
}

// serverURLValue is a flag.Value that validates the URL scheme on Set.
type serverURLValue struct {
	target *string
}

func (v *serverURLValue) String() string {
	if v.target == nil {
		return ""
	}
	return *v.target
}

func (v *serverURLValue) Set(s string) error {
	if err := validateServerURL(s); err != nil {
		return err
	}
	*v.target = s
	return nil
}

func addServerFlag(fs *flag.FlagSet, serverURL *string) {
	fs.Var(&serverURLValue{target: serverURL}, "server", "PeerClaw server URL")
}

func addOutputFlag(fs *flag.FlagSet) {
	fs.StringVar(&outputFormat, "output", "table", "Output format (table or json)")
}

// reorderArgs moves positional arguments after all flags so that Go's
// flag.Parse (which stops at the first non-flag argument) processes
// all flags correctly regardless of argument order.
//
// Example: ["agent-id", "--loop", "--status", "online"]
// becomes: ["--loop", "--status", "online", "agent-id"]
func reorderArgs(fs *flag.FlagSet, args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}

		flags = append(flags, arg)

		// Check if value is embedded (-flag=value).
		name := strings.TrimLeft(arg, "-")
		if strings.Contains(name, "=") {
			continue
		}

		// Look up the flag to see if it takes a value.
		f := fs.Lookup(name)
		if f == nil {
			continue // unknown flag — let flag.Parse report the error
		}

		// Boolean flags don't consume the next argument.
		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			continue
		}

		// Non-boolean flag — next arg is the value.
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...)
}
