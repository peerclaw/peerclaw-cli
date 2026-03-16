package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunContactRequests handles the "agent contact-requests" subcommand.
func RunContactRequests(args []string, serverURL string) int {
	if len(args) < 1 {
		printContactReqUsage()
		return 1
	}

	switch args[0] {
	case "send":
		return runContactReqSend(args[1:], serverURL)
	case "list":
		return runContactReqList(args[1:], serverURL)
	case "approve":
		return runContactReqApprove(args[1:], serverURL)
	case "reject":
		return runContactReqReject(args[1:], serverURL)
	case "help", "-h":
		printContactReqUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown contact-requests command: %s\n", args[0])
		printContactReqUsage()
		return 1
	}
}

func printContactReqUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw agent contact-requests <subcommand> [options]

Subcommands:
  send       Send a contact request to another agent
  list       List contact requests (incoming or sent)
  approve    Approve a contact request
  reject     Reject a contact request
`)
}

func runContactReqSend(args []string, serverURL string) int {
	fs := flag.NewFlagSet("contact-requests send", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	target := fs.String("target", "", "Target agent ID (required)")
	message := fs.String("message", "", "Optional message")
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contact-requests send <agent-id> --target <target-id>\n")
		return 1
	}
	if *target == "" {
		fmt.Fprintf(os.Stderr, "Error: --target is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	req, err := c.SendContactRequest(context.Background(), agentID, map[string]string{
		"target_agent_id": *target,
		"message":         *message,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Contact request sent: %s\n", req.ID)
	PrintJSON(req)
	return 0
}

func runContactReqList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("contact-requests list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	direction := fs.String("direction", "incoming", "Direction: incoming or sent")
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contact-requests list <agent-id> [--direction incoming|sent]\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)

	var resp *client.ListContactRequestsResponse
	var err error
	if *direction == "sent" {
		resp, err = c.ListSentContactRequests(context.Background(), agentID)
	} else {
		resp, err = c.ListIncomingContactRequests(context.Background(), agentID)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"ID", "FROM", "TO", "STATUS", "MESSAGE", "CREATED"}
	var rows [][]string
	for _, r := range resp.Requests {
		rows = append(rows, []string{
			r.ID,
			r.FromAgentID,
			r.ToAgentID,
			r.Status,
			r.Message,
			r.CreatedAt,
		})
	}

	fmt.Fprintf(os.Stderr, "Total: %d requests\n", len(resp.Requests))
	PrintAuto(headers, rows, resp)
	return 0
}

func runContactReqApprove(args []string, serverURL string) int {
	fs := flag.NewFlagSet("contact-requests approve", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	requestID := fs.String("request", "", "Request ID (required)")
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contact-requests approve <agent-id> --request <request-id>\n")
		return 1
	}
	if *requestID == "" {
		fmt.Fprintf(os.Stderr, "Error: --request is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	if err := c.UpdateContactRequest(context.Background(), agentID, *requestID, map[string]string{
		"action": "approve",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Contact request approved: %s\n", *requestID)
	return 0
}

func runContactReqReject(args []string, serverURL string) int {
	fs := flag.NewFlagSet("contact-requests reject", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	requestID := fs.String("request", "", "Request ID (required)")
	reason := fs.String("reason", "", "Rejection reason")
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contact-requests reject <agent-id> --request <request-id>\n")
		return 1
	}
	if *requestID == "" {
		fmt.Fprintf(os.Stderr, "Error: --request is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	if err := c.UpdateContactRequest(context.Background(), agentID, *requestID, map[string]string{
		"action": "reject",
		"reason": *reason,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Contact request rejected: %s\n", *requestID)
	return 0
}
