package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

func runAgentHeartbeat(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent heartbeat", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	status := fs.String("status", "online", "Agent status (online, degraded)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent heartbeat <agent-id> [--status online|degraded]\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.Heartbeat(context.Background(), agentID, client.HeartbeatRequest{
		Status: *status,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Heartbeat sent for agent %s\n", agentID)
	fmt.Fprintf(os.Stderr, "  Next deadline: %s\n", resp.NextDeadline)
	return 0
}
