package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	agent "github.com/peerclaw/peerclaw-agent"
)

// RunTransfer handles the "transfer" subcommand.
func RunTransfer(args []string, serverURL string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw transfer <status> [options]\n")
		return 1
	}

	switch args[0] {
	case "status":
		return runTransferStatus(args[1:], serverURL)
	default:
		fmt.Fprintf(os.Stderr, "unknown transfer subcommand: %s\n", args[0])
		return 1
	}
}

func runTransferStatus(args []string, serverURL string) int {
	fs := flag.NewFlagSet("transfer status", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	transferID := fs.String("transfer-id", "", "Specific transfer ID to show")
	keypairPath := fs.String("keypair", "", "Path to Ed25519 keypair file")
	trustStorePath := fs.String("trust-store", "", "Path to trust store file")
	fs.Parse(args)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create agent to query transfers.
	a, err := agent.New(agent.Options{
		Name:           "peerclaw-cli",
		ServerURL:      serverURL,
		KeypairPath:    *keypairPath,
		TrustStorePath: *trustStorePath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		return 1
	}

	if err := a.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting agent: %v\n", err)
		return 1
	}
	defer a.Stop(context.Background())

	if *transferID != "" {
		ti, ok := a.GetTransfer(*transferID)
		if !ok {
			fmt.Fprintf(os.Stderr, "Transfer not found: %s\n", *transferID)
			return 1
		}
		PrintJSON(ti)
		return 0
	}

	transfers := a.ListTransfers()
	if len(transfers) == 0 {
		fmt.Println("No active transfers")
		return 0
	}

	headers := []string{"ID", "FILE", "PEER", "DIR", "STATE", "PROGRESS"}
	var rows [][]string
	for _, t := range transfers {
		rows = append(rows, []string{
			t.FileID[:8] + "...",
			t.FileName,
			truncateID(t.PeerID),
			string(t.Direction),
			string(t.State),
			fmt.Sprintf("%.1f%%", t.Progress*100),
		})
	}

	PrintAuto(headers, rows, transfers)
	return 0
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12] + "..."
	}
	return id
}
