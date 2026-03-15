package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peerclaw/peerclaw-cli/internal/client"
	"github.com/peerclaw/peerclaw-core/identity"
)

func runAgentClaim(args []string, serverURL string) int {
	fs := flag.NewFlagSet("agent claim", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	token := fs.String("token", "", "Claim token (e.g., PCW-XXXX-XXXX) (required)")
	name := fs.String("name", "", "Agent name (optional — defaults to name set in token)")
	keypairPath := fs.String("keypair", "./agent.key", "Path to save/load Ed25519 keypair")
	capabilities := fs.String("capabilities", "", "Comma-separated capabilities (optional)")
	protocols := fs.String("protocols", "", "Comma-separated protocols (optional)")
	endpointURL := fs.String("url", "", "Agent endpoint URL (optional)")
	fs.Parse(args)

	if *token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: peerclaw agent claim --token <PCW-XXXX-XXXX>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
		return 1
	}

	// Load or generate Ed25519 keypair.
	kp, err := identity.LoadKeypair(*keypairPath)
	if err != nil {
		kp, err = identity.GenerateKeypair()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating keypair: %v\n", err)
			return 1
		}
		if err := identity.SaveKeypair(kp, *keypairPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving keypair: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "Generated new keypair: %s\n", *keypairPath)
	} else {
		fmt.Fprintf(os.Stderr, "Loaded existing keypair: %s\n", *keypairPath)
	}

	// Sign the token to prove key ownership.
	sig, err := identity.Sign(kp.PrivateKey, []byte(*token))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing token: %v\n", err)
		return 1
	}

	var caps []string
	if *capabilities != "" {
		caps = strings.Split(*capabilities, ",")
	}
	var protos []string
	if *protocols != "" {
		protos = strings.Split(*protocols, ",")
	}

	c := client.New(serverURL)
	card, err := c.ClaimAgent(context.Background(), client.ClaimRequest{
		Token:        *token,
		Name:         *name, // empty string is fine — server uses token metadata
		PublicKey:    kp.PublicKeyString(),
		Capabilities: caps,
		Protocols:    protos,
		Endpoint:     client.EndpointReq{URL: *endpointURL},
		Signature:    sig,
		Metadata:     map[string]string{"sdk_version": Version},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Save agent ID, keypair path, and server to config for future use.
	cfg, _ := loadCLIConfig()
	cfg.AgentID = card.ID
	cfg.Keypair = *keypairPath
	if cfg.Server == "" || cfg.Server == defaultServer {
		cfg.Server = serverURL
	}
	if err := saveCLIConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "\nAgent registered successfully!\n")
	fmt.Fprintf(os.Stderr, "  ID:         %s\n", card.ID)
	fmt.Fprintf(os.Stderr, "  Name:       %s\n", card.Name)
	fmt.Fprintf(os.Stderr, "  Public Key: %s\n", kp.PublicKeyString())
	fmt.Fprintf(os.Stderr, "  Keypair:    %s\n", *keypairPath)
	fmt.Fprintf(os.Stderr, "  Config:     %s\n\n", configPath())
	fmt.Fprintf(os.Stderr, "Next steps:\n")
	fmt.Fprintf(os.Stderr, "  peerclaw agent get %s          # verify registration\n", card.ID)
	fmt.Fprintf(os.Stderr, "  peerclaw mcp serve                       # run as MCP tool server (auto-heartbeat)\n")
	fmt.Fprintf(os.Stderr, "  peerclaw invoke <agent-id> --message \"Hello\"  # talk to other agents\n\n")
	fmt.Fprintf(os.Stderr, "Keep %s safe — it proves your agent's identity.\n", *keypairPath)
	fmt.Fprintf(os.Stderr, "Docs: https://github.com/peerclaw/peerclaw/blob/main/docs/GUIDE.md\n")

	return 0
}
