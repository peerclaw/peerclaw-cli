package cmd

import (
	"flag"
	"fmt"
	"net/url"
	"os"
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
