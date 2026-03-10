package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunReputation handles the "reputation" subcommand.
func RunReputation(args []string, serverURL string) int {
	if len(args) < 1 {
		printReputationUsage()
		return 1
	}

	switch args[0] {
	case "show":
		return runReputationShow(args[1:], serverURL)
	case "list":
		return runReputationList(args[1:], serverURL)
	case "help", "-h":
		printReputationUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown reputation command: %s\n", args[0])
		printReputationUsage()
		return 1
	}
}

func printReputationUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw reputation <subcommand> [options]

Subcommands:
  show <agent-id>   Show reputation score and history for an agent
  list              List agents with their reputation scores
`)
}

func runReputationShow(args []string, serverURL string) int {
	fs := flag.NewFlagSet("reputation show", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	limit := fs.Int("limit", 10, "Number of recent reputation events to show")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw reputation show <agent-id>\n")
		return 1
	}

	agentID := fs.Arg(0)
	c := client.New(serverURL)

	// Get the agent's public profile for the current score.
	profile, err := c.GetDirectoryAgent(context.Background(), agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching agent profile: %v\n", err)
		return 1
	}

	// Get reputation event history.
	history, err := c.GetReputationHistory(context.Background(), agentID, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching reputation history: %v\n", err)
		return 1
	}

	if outputFormat == "json" {
		PrintJSON(map[string]any{
			"agent_id":          profile.ID,
			"name":              profile.Name,
			"reputation_score":  profile.ReputationScore,
			"reputation_events": profile.ReputationEvents,
			"verified":          profile.Verified,
			"trusted":           profile.Trusted,
			"events":            history.Events,
		})
		return 0
	}

	// Table output.
	fmt.Printf("Agent:            %s\n", profile.ID)
	fmt.Printf("Name:             %s\n", profile.Name)
	fmt.Printf("Reputation Score: %.4f\n", profile.ReputationScore)
	fmt.Printf("Total Events:     %d\n", profile.ReputationEvents)
	fmt.Printf("Verified:         %v\n", profile.Verified)
	fmt.Printf("Trusted:          %v\n", profile.Trusted)

	if len(history.Events) > 0 {
		fmt.Printf("\nRecent Events (last %d):\n", *limit)
		headers := []string{"EVENT TYPE", "WEIGHT", "SCORE AFTER", "CREATED AT"}
		var rows [][]string
		for _, e := range history.Events {
			rows = append(rows, []string{
				e.EventType,
				fmt.Sprintf("%+.2f", e.Weight),
				fmt.Sprintf("%.4f", e.ScoreAfter),
				e.CreatedAt,
			})
		}
		PrintTable(headers, rows)
	} else {
		fmt.Println("\nNo reputation events recorded yet.")
	}

	return 0
}

func runReputationList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("reputation list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	protocol := fs.String("protocol", "", "Filter by protocol")
	capability := fs.String("capability", "", "Filter by capability")
	status := fs.String("status", "", "Filter by status")
	fs.Parse(args)

	c := client.New(serverURL)
	resp, err := c.ListDirectory(context.Background(), client.ListAgentsOptions{
		Protocol:   *protocol,
		Capability: *capability,
		Status:     *status,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"ID", "NAME", "SCORE", "EVENTS", "VERIFIED", "STATUS", "CAPABILITIES"}
	var rows [][]string
	for _, a := range resp.Agents {
		verified := ""
		if a.Verified {
			verified = "yes"
		}
		rows = append(rows, []string{
			a.ID,
			a.Name,
			fmt.Sprintf("%.4f", a.ReputationScore),
			fmt.Sprintf("%d", a.ReputationEvents),
			verified,
			a.Status,
			strings.Join(a.Capabilities, ","),
		})
	}

	fmt.Fprintf(os.Stderr, "Total: %d agents\n", resp.TotalCount)
	PrintAuto(headers, rows, resp)
	return 0
}
