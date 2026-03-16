package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunInbox handles the "inbox" command.
func RunInbox(args []string, serverURL string) int {
	if len(args) < 1 {
		printInboxUsage()
		return 1
	}

	switch args[0] {
	case "request":
		return runInboxRequest(args[1:], serverURL)
	case "status":
		return runInboxStatus(args[1:], serverURL)
	case "list":
		return runInboxList(args[1:], serverURL)
	case "help", "-h":
		printInboxUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown inbox command: %s\n", args[0])
		printInboxUsage()
		return 1
	}
}

func printInboxUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw inbox <subcommand> [options]

Subcommands:
  request    Submit access request to an agent
  status     Check access request status for an agent
  list       List all my access requests

Global options:
  --token    JWT token (or PEERCLAW_TOKEN env)
  --server   PeerClaw server URL
`)
}

func runInboxRequest(args []string, serverURL string) int {
	fs := flag.NewFlagSet("inbox request", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	var token string
	addTokenFlag(fs, &token)
	message := fs.String("message", "", "Access request message")
	fs.StringVar(message, "m", "", "Access request message (shorthand)")
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw inbox request <agent-id> [options]\n")
		return 1
	}

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.SubmitAccessRequest(context.Background(), agentID, *message, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Access request submitted: %s (status: %s)\n", resp.ID, resp.Status)
	PrintJSON(resp)
	return 0
}

func runInboxStatus(args []string, serverURL string) int {
	fs := flag.NewFlagSet("inbox status", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	var token string
	addTokenFlag(fs, &token)
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw inbox status <agent-id>\n")
		return 1
	}

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.GetMyAccessRequest(context.Background(), agentID, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	PrintJSON(resp)
	return 0
}

func runInboxList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("inbox list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	var token string
	addTokenFlag(fs, &token)
	fs.Parse(args)

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	c := client.New(serverURL)
	requests, err := c.ListMyAccessRequests(context.Background(), token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"ID", "AGENT", "STATUS", "MESSAGE", "CREATED"}
	var rows [][]string
	for _, r := range requests {
		msg := r.Message
		if len(msg) > 40 {
			msg = msg[:40] + "..."
		}
		rows = append(rows, []string{r.ID, r.AgentID, r.Status, msg, r.CreatedAt})
	}

	fmt.Fprintf(os.Stderr, "Total: %d access requests\n", len(requests))
	PrintAuto(headers, rows, requests)
	return 0
}
