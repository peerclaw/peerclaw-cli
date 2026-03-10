package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunAgent handles the "agent" subcommand.
func RunAgent(args []string, serverURL string) int {
	if len(args) < 1 {
		printAgentUsage()
		return 1
	}

	switch args[0] {
	case "list":
		return runAgentList(args[1:], serverURL)
	case "get":
		return runAgentGet(args[1:], serverURL)
	case "register":
		return runAgentRegister(args[1:], serverURL)
	case "claim":
		return runAgentClaim(args[1:], serverURL)
	case "delete":
		return runAgentDelete(args[1:], serverURL)
	case "discover":
		return runAgentDiscover(args[1:], serverURL)
	case "heartbeat":
		return runAgentHeartbeat(args[1:], serverURL)
	case "verify":
		return runAgentVerify(args[1:], serverURL)
	case "contacts":
		return RunContacts(args[1:], serverURL)
	case "help", "-h":
		printAgentUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown agent command: %s\n", args[0])
		printAgentUsage()
		return 1
	}
}

func printAgentUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw agent <subcommand> [options]

Subcommands:
  list       List registered agents
  get        Get agent details
  register   Register a new agent
  claim      Register an agent using a claim token
  delete     Delete an agent
  discover   Discover agents by capabilities
  heartbeat  Send heartbeat for an agent
  verify     Verify an agent's endpoint
  contacts   Manage agent contacts whitelist
`)
}

func runAgentList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	protocol := fs.String("protocol", "", "Filter by protocol")
	capability := fs.String("capability", "", "Filter by capability")
	status := fs.String("status", "", "Filter by status")
	fs.Parse(args)

	c := client.New(serverURL)
	resp, err := c.ListAgents(context.Background(), client.ListAgentsOptions{
		Protocol:   *protocol,
		Capability: *capability,
		Status:     *status,
	})
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

	fmt.Fprintf(os.Stderr, "Total: %d agents\n", resp.TotalCount)
	PrintAuto(headers, rows, resp)
	return 0
}

func runAgentGet(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent get", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent get <agent-id>\n")
		return 1
	}

	c := client.New(serverURL)
	card, err := c.GetAgent(context.Background(), fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	PrintJSON(card)
	return 0
}

func runAgentRegister(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent register", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	name := fs.String("name", "", "Agent name (required)")
	description := fs.String("description", "", "Agent description")
	version := fs.String("version", "", "Agent version")
	protocols := fs.String("protocols", "a2a", "Comma-separated protocols")
	capabilities := fs.String("capabilities", "", "Comma-separated capabilities")
	endpointURL := fs.String("url", "", "Agent endpoint URL (required)")
	fs.Parse(args)

	if *name == "" || *endpointURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --name and --url are required\n")
		fs.Usage()
		return 1
	}

	var caps []string
	if *capabilities != "" {
		caps = strings.Split(*capabilities, ",")
	}

	c := client.New(serverURL)
	card, err := c.RegisterAgent(context.Background(), client.RegisterRequest{
		Name:         *name,
		Description:  *description,
		Version:      *version,
		Capabilities: caps,
		Protocols:    strings.Split(*protocols, ","),
		Endpoint:     client.EndpointReq{URL: *endpointURL},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Agent registered: %s\n", card.ID)
	PrintJSON(card)
	return 0
}

func runAgentDelete(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent delete", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent delete <agent-id>\n")
		return 1
	}

	c := client.New(serverURL)
	if err := c.DeleteAgent(context.Background(), fs.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Agent deleted: %s\n", fs.Arg(0))
	return 0
}
