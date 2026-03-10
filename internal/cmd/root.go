package cmd

import (
	"flag"
	"fmt"
	"os"
)

const defaultServer = "http://localhost:8080"

// Run executes the CLI with the given arguments.
func Run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 1
	}

	// Global flags.
	serverURL := os.Getenv("PEERCLAW_SERVER")
	if serverURL == "" {
		serverURL = defaultServer
	}

	switch args[0] {
	case "agent":
		return RunAgent(args[1:], serverURL)
	case "send":
		return RunSend(args[1:], serverURL)
	case "health":
		return RunHealth(args[1:], serverURL)
	case "config":
		return RunConfig(args[1:])
	case "dht":
		return RunDHT(args[1:], serverURL)
	case "federation":
		return RunFederation(args[1:], serverURL)
	case "reputation":
		return RunReputation(args[1:], serverURL)
	case "identity":
		return RunIdentity(args[1:], serverURL)
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version":
		fmt.Println("peerclaw version 0.5.0")
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
  agent       Manage agents (list, get, register, claim, discover, heartbeat, verify)
  send        Send a message through the bridge
  health      Check server health
  config      Manage CLI configuration
  dht         DHT operations (bootstrap, lookup)
  federation  Federation management (status, peers)
  reputation  Reputation scores (show, list)
  identity    Identity anchoring (anchor, verify)
  version     Print version

Environment:
  PEERCLAW_SERVER   Server URL (default: %s)

Use "peerclaw <command> -h" for more information.
`, defaultServer)
}

func addServerFlag(fs *flag.FlagSet, serverURL *string) {
	fs.StringVar(serverURL, "server", *serverURL, "PeerClaw server URL")
}

func addOutputFlag(fs *flag.FlagSet) {
	fs.StringVar(&outputFormat, "output", "table", "Output format (table or json)")
}
