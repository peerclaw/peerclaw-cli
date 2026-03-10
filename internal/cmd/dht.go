package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunDHT handles the "dht" subcommand.
func RunDHT(args []string, serverURL string) int {
	if len(args) < 1 {
		printDHTUsage()
		return 1
	}

	switch args[0] {
	case "bootstrap":
		return runDHTBootstrap(args[1:], serverURL)
	case "lookup":
		return runDHTLookup(args[1:], serverURL)
	case "help", "-h":
		printDHTUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown dht command: %s\n", args[0])
		printDHTUsage()
		return 1
	}
}

func printDHTUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw dht <subcommand> [options]

Subcommands:
  bootstrap   Bootstrap the DHT with seed nodes
  lookup      Look up an agent by public key in the DHT
`)
}

func runDHTBootstrap(args []string, serverURL string) int {
	fs := flag.NewFlagSet("dht bootstrap", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	seeds := fs.String("seeds", "", "Comma-separated seed node addresses")
	relays := fs.String("relays", "", "Comma-separated Nostr relay URLs")
	fs.Parse(args)

	fmt.Fprintf(os.Stderr, "Checking server connectivity at %s ...\n", serverURL)

	c := client.New(serverURL)
	resp, err := c.Health(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to reach server: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Server status: %s\n", resp.Status)
	if resp.RegisteredAgents > 0 {
		fmt.Fprintf(os.Stderr, "Registered agents: %d\n", resp.RegisteredAgents)
	}
	if resp.ConnectedAgents > 0 {
		fmt.Fprintf(os.Stderr, "Connected agents: %d\n", resp.ConnectedAgents)
	}
	if *seeds != "" {
		fmt.Fprintf(os.Stderr, "Seeds: %s\n", *seeds)
	}
	if *relays != "" {
		fmt.Fprintf(os.Stderr, "Relays: %s\n", *relays)
	}

	fmt.Println("DHT bootstrapped via server — server is reachable and healthy.")
	fmt.Println("Note: Full peer-to-peer DHT requires the Agent SDK. The CLI uses the server's")
	fmt.Println("centralized directory for discovery. Use 'peerclaw dht lookup' to find agents.")
	return 0
}

func runDHTLookup(args []string, serverURL string) int {
	fs := flag.NewFlagSet("dht lookup", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw dht lookup <key>\n")
		fmt.Fprintf(os.Stderr, "\nThe key is used as a capability query to discover matching agents.\n")
		return 1
	}

	key := fs.Arg(0)
	fmt.Fprintf(os.Stderr, "Looking up agents matching: %s\n", key)

	c := client.New(serverURL)
	resp, err := c.Discover(context.Background(), []string{key}, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"ID", "NAME", "STATUS", "PROTOCOLS", "CAPABILITIES"}
	var rows [][]string
	for _, a := range resp.Agents {
		protos := make([]string, len(a.Protocols))
		for i, p := range a.Protocols {
			protos[i] = string(p)
		}
		rows = append(rows, []string{
			a.ID,
			a.Name,
			string(a.Status),
			strings.Join(protos, ","),
			strings.Join(a.Capabilities, ","),
		})
	}

	fmt.Fprintf(os.Stderr, "Found %d matching agents\n", len(resp.Agents))
	PrintAuto(headers, rows, resp)
	return 0
}
