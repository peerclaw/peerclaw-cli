package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

func runAgentDiscover(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent discover", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	capabilities := fs.String("capabilities", "", "Comma-separated capabilities to search for (required)")
	protocol := fs.String("protocol", "", "Filter by protocol")
	fs.Parse(args)

	if *capabilities == "" {
		fmt.Fprintf(os.Stderr, "Error: --capabilities is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent discover --capabilities <cap1,cap2> [--protocol <proto>]\n")
		return 1
	}

	caps := strings.Split(*capabilities, ",")
	c := client.New(serverURL)
	resp, err := c.Discover(context.Background(), caps, *protocol)
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
