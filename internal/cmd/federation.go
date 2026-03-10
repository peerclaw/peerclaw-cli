package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunFederation handles the "federation" subcommand.
func RunFederation(args []string, serverURL string) int {
	if len(args) < 1 {
		printFederationUsage()
		return 1
	}

	switch args[0] {
	case "status":
		return runFederationStatus(args[1:], serverURL)
	case "peers":
		return runFederationPeers(args[1:], serverURL)
	case "help", "-h":
		printFederationUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown federation command: %s\n", args[0])
		printFederationUsage()
		return 1
	}
}

func printFederationUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw federation <subcommand> [options]

Subcommands:
  status    Show federation status
  peers     List federated peers
`)
}

func runFederationStatus(args []string, serverURL string) int {
	fs := flag.NewFlagSet("federation status", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	fs.Parse(args)

	c := client.New(serverURL)
	health, err := c.Health(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Server Status: %s\n", health.Status)
	if fed, ok := health.Components["federation"]; ok {
		fmt.Printf("Federation: %s\n", fed)
	} else {
		fmt.Println("Federation: not configured")
	}
	return 0
}

func runFederationPeers(args []string, serverURL string) int {
	fs := flag.NewFlagSet("federation peers", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	fs.Parse(args)

	// Federation peer info would come from a dedicated API endpoint.
	// For now, we show a placeholder using the health endpoint.
	c := client.New(serverURL)
	_, err := c.Health(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("Federation peers (use server config to view peer details):")
	fmt.Println("  (query /api/v1/health for federation status)")
	return 0
}
