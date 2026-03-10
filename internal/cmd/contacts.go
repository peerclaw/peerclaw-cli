package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunContacts handles the "agent contacts" subcommand.
func RunContacts(args []string, serverURL string) int {
	if len(args) < 1 {
		printContactsUsage()
		return 1
	}

	switch args[0] {
	case "list":
		return runContactsList(args[1:], serverURL)
	case "add":
		return runContactsAdd(args[1:], serverURL)
	case "remove":
		return runContactsRemove(args[1:], serverURL)
	case "help", "-h":
		printContactsUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown contacts command: %s\n", args[0])
		printContactsUsage()
		return 1
	}
}

func printContactsUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw agent contacts <subcommand> [options]

Manage the agent's contacts whitelist. Only agents in the whitelist
can send messages to this agent.

Subcommands:
  list     List contacts for an agent
  add      Add a contact to the whitelist
  remove   Remove a contact from the whitelist
`)
}

func runContactsList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent contacts list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contacts list <agent-id>\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.ListContacts(context.Background(), agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"CONTACT_AGENT_ID", "ALIAS", "CREATED_AT"}
	var rows [][]string
	for _, ct := range resp.Contacts {
		rows = append(rows, []string{
			ct.ContactAgentID,
			ct.Alias,
			ct.CreatedAt,
		})
	}

	fmt.Fprintf(os.Stderr, "Total: %d contacts\n", len(resp.Contacts))
	PrintAuto(headers, rows, resp)
	return 0
}

func runContactsAdd(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent contacts add", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	contact := fs.String("contact", "", "Contact agent ID to add (required)")
	alias := fs.String("alias", "", "Optional alias for the contact")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contacts add <agent-id> --contact <contact-agent-id> [--alias \"name\"]\n")
		return 1
	}
	if *contact == "" {
		fmt.Fprintf(os.Stderr, "Error: --contact is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	resp, err := c.AddContact(context.Background(), agentID, client.AddContactRequest{
		ContactAgentID: *contact,
		Alias:          *alias,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Contact added: %s -> %s\n", agentID, *contact)
	PrintJSON(resp)
	return 0
}

func runContactsRemove(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent contacts remove", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	contact := fs.String("contact", "", "Contact agent ID to remove (required)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent contacts remove <agent-id> --contact <contact-agent-id>\n")
		return 1
	}
	if *contact == "" {
		fmt.Fprintf(os.Stderr, "Error: --contact is required\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)
	if err := c.RemoveContact(context.Background(), agentID, *contact); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Contact removed: %s -/-> %s\n", agentID, *contact)
	return 0
}
