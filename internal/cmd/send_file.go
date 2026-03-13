package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	agent "github.com/peerclaw/peerclaw-agent"
)

// RunSendFile handles the "send-file" subcommand.
func RunSendFile(args []string, serverURL string) int {
	fs := flag.NewFlagSet("send-file", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	to := fs.String("to", "", "Destination agent ID (required)")
	filePath := fs.String("file", "", "Path to file to send (required)")
	keypairPath := fs.String("keypair", "", "Path to Ed25519 keypair file")
	trustStorePath := fs.String("trust-store", "", "Path to trust store file")
	noRegister := fs.Bool("no-register", false, "Skip server registration (reuse existing identity)")
	fs.Parse(args)

	if *to == "" || *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: --to and --file are required\n")
		fs.Usage()
		return 1
	}

	// Verify file exists.
	info, err := os.Stat(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create agent for P2P transfer.
	a, err := agent.New(agent.Options{
		Name:             "peerclaw-cli-sender",
		ServerURL:        serverURL,
		Capabilities:     []string{"file_transfer"},
		KeypairPath:      *keypairPath,
		TrustStorePath:   *trustStorePath,
		SkipRegistration: *noRegister,
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

	fmt.Printf("Sending %s (%d bytes) to %s...\n", info.Name(), info.Size(), *to)

	fileID, err := a.SendFile(ctx, *to, *filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Transfer initiated: %s\n", fileID)

	// Poll for progress.
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "\nTransfer cancelled\n")
			_ = a.CancelTransfer(fileID)
			return 1
		case <-ticker.C:
			ti, ok := a.GetTransfer(fileID)
			if !ok {
				continue
			}
			fmt.Printf("\r  Progress: %.1f%% (%d bytes sent)", ti.Progress*100, ti.BytesSent)

			switch ti.State {
			case "done":
				fmt.Printf("\nTransfer complete!\n")
				return 0
			case "failed":
				fmt.Fprintf(os.Stderr, "\nTransfer failed: %s\n", ti.Error)
				return 1
			case "cancelled":
				fmt.Fprintf(os.Stderr, "\nTransfer cancelled\n")
				return 1
			}
		}
	}
}
