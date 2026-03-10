package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunSend handles the "send" subcommand.
func RunSend(args []string, serverURL string) int {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	from := fs.String("from", "", "Source agent ID (required)")
	to := fs.String("to", "", "Destination agent ID (required)")
	protocol := fs.String("protocol", "", "Protocol (a2a, mcp, acp)")
	payload := fs.String("payload", "", "Message payload (required)")
	fs.Parse(args)

	if *from == "" || *to == "" || *payload == "" {
		fmt.Fprintf(os.Stderr, "Error: --from, --to, and --payload are required\n")
		fs.Usage()
		return 1
	}

	c := client.New(serverURL)
	resp, err := c.Send(context.Background(), client.SendRequest{
		Source:      *from,
		Destination: *to,
		Protocol:    *protocol,
		Payload:     *payload,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	PrintJSON(resp)
	return 0
}
