package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunInvoke handles the "invoke" command.
func RunInvoke(args []string, serverURL string) int {
	fs := flag.NewFlagSet("invoke", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	message := fs.String("message", "", "Message to send (required)")
	fs.StringVar(message, "m", "", "Message to send (shorthand)")
	protocol := fs.String("protocol", "", "Protocol (default: auto)")
	stream := fs.Bool("stream", false, "Enable streaming output (SSE)")
	sessionID := fs.String("session-id", "", "Session ID for conversation continuity")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw invoke <agent-id> [options]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fs.PrintDefaults()
		return 1
	}

	agentID := fs.Arg(0)

	if *message == "" {
		fmt.Fprintf(os.Stderr, "Error: --message (-m) is required\n")
		return 1
	}

	req := client.InvokeRequest{
		Message:   *message,
		Protocol:  *protocol,
		SessionID: *sessionID,
	}

	c := client.New(serverURL)
	ctx := context.Background()

	if *stream {
		return runInvokeStream(ctx, c, agentID, req)
	}
	return runInvokeSync(ctx, c, agentID, req)
}

func runInvokeSync(ctx context.Context, c *client.Client, agentID string, req client.InvokeRequest) int {
	resp, err := c.Invoke(ctx, agentID, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	PrintJSON(resp)
	return 0
}

func runInvokeStream(ctx context.Context, c *client.Client, agentID string, req client.InvokeRequest) int {
	scanner, body, err := c.InvokeStream(ctx, agentID, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer body.Close()

	for {
		ev := client.ParseSSE(scanner)
		if ev == nil {
			break
		}
		switch ev.Event {
		case "message":
			var msg struct {
				Content string `json:"content"`
			}
			if json.Unmarshal([]byte(ev.Data), &msg) == nil {
				fmt.Print(msg.Content)
			}
		case "done":
			fmt.Println()
			return 0
		case "error":
			fmt.Fprintf(os.Stderr, "\nError: %s\n", ev.Data)
			return 1
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "\nStream error: %v\n", err)
		return 1
	}
	fmt.Println()
	return 0
}
