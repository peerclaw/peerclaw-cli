package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunHealth handles the "health" subcommand.
func RunHealth(args []string, serverURL string) int {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	fs.Parse(args)

	c := client.New(serverURL)
	resp, err := c.Health(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if outputFormat == "json" {
		PrintJSON(resp)
	} else {
		fmt.Printf("Status: %s\n", resp.Status)
		if resp.Components != nil {
			for k, v := range resp.Components {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		if resp.ConnectedAgents > 0 || resp.RegisteredAgents > 0 {
			fmt.Printf("Connected Agents: %d\n", resp.ConnectedAgents)
			fmt.Printf("Registered Agents: %d\n", resp.RegisteredAgents)
		}
	}
	return 0
}
