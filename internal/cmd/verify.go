package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

func runAgentVerify(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent verify", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent verify <agent-id>\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.VerifyEndpoint(context.Background(), agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Agent %s verification: %s\n", agentID, resp.Status)
	if resp.Challenge != "" {
		fmt.Fprintf(os.Stderr, "  Challenge: %s\n", resp.Challenge)
	}
	return 0
}
