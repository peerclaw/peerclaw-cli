package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/peerclaw/peerclaw-cli/internal/acpserve"
)

// RunACP runs the ACP command.
func RunACP(args []string, serverURL string) int {
	if len(args) < 1 {
		printACPUsage()
		return 1
	}
	switch args[0] {
	case "serve":
		return runACPServe(args[1:], serverURL)
	default:
		fmt.Fprintf(os.Stderr, "unknown acp command: %s\n\n", args[0])
		printACPUsage()
		return 1
	}
}

func runACPServe(args []string, serverURL string) int {
	fs := flag.NewFlagSet("acp serve", flag.ContinueOnError)
	addServerFlag(fs, &serverURL)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	server := acpserve.New(serverURL, os.Stdin, os.Stdout)

	fmt.Fprintf(os.Stderr, "ACP stdio bridge started (server: %s)\n", serverURL)

	if err := server.Run(ctx); err != nil {
		if ctx.Err() == nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	return 0
}

func printACPUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw acp <command> [options]

Commands:
  serve     Start ACP stdio bridge (ndJSON on stdin/stdout)

Options (serve):
  --server      PeerClaw server URL (default: %s)

Examples:
  peerclaw acp serve
  peerclaw acp serve --server http://localhost:8080
  echo '{"id":"1","method":"ping"}' | peerclaw acp serve
`, defaultServer)
}
