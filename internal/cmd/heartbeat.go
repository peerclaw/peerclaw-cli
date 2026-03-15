package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

func runAgentHeartbeat(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent heartbeat", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	status := fs.String("status", "online", "Agent status (online, degraded)")
	loop := fs.Bool("loop", false, "Send heartbeats continuously")
	interval := fs.Duration("interval", 30*time.Second, "Heartbeat interval (used with --loop)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent heartbeat <agent-id> [--status online|degraded] [--loop] [--interval 30s]\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)

	sendOne := func() error {
		resp, err := c.Heartbeat(context.Background(), agentID, client.HeartbeatRequest{
			Status:   *status,
			Metadata: map[string]string{"sdk_version": Version},
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Heartbeat sent for agent %s (next deadline: %s)\n", agentID, resp.NextDeadline)
		return nil
	}

	// Single heartbeat mode.
	if !*loop {
		if err := sendOne(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	// Continuous heartbeat mode.
	fmt.Fprintf(os.Stderr, "Starting heartbeat loop for agent %s (interval: %s)\n", agentID, *interval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Send first heartbeat immediately.
	if err := sendOne(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "\nHeartbeat loop stopped.\n")
			return 0
		case <-ticker.C:
			if err := sendOne(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}
	}
}
