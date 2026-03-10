package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/peerclaw/peerclaw-cli/internal/client"
	"github.com/peerclaw/peerclaw-core/identity"
)

// RunIdentity handles the "identity" subcommand.
func RunIdentity(args []string, serverURL string) int {
	if len(args) < 1 {
		printIdentityUsage()
		return 1
	}

	switch args[0] {
	case "anchor":
		return runIdentityAnchor(args[1:], serverURL)
	case "verify":
		return runIdentityVerify(args[1:], serverURL)
	case "help", "-h":
		printIdentityUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown identity command: %s\n", args[0])
		printIdentityUsage()
		return 1
	}
}

func printIdentityUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw identity <subcommand> [options]

Subcommands:
  anchor           Generate a Nostr identity anchor event
  verify <agent-id> Verify an agent's endpoint via the server
`)
}

// nostrEvent is a simplified Nostr event structure for identity anchoring.
type nostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

func runIdentityAnchor(args []string, serverURL string) int {
	fs := flag.NewFlagSet("identity anchor", flag.ExitOnError)
	keypairPath := fs.String("keypair", "", "Path to keypair seed file (required)")
	relays := fs.String("relays", "wss://relay.damus.io,wss://nos.lol", "Comma-separated Nostr relay URLs for publishing")
	agentName := fs.String("name", "", "Agent name to include in the anchor")
	fs.Parse(args)

	if *keypairPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --keypair is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: peerclaw identity anchor --keypair <path> [--name <agent-name>] [--relays <relay-urls>]\n")
		return 1
	}

	// Load the keypair.
	kp, err := identity.LoadKeypair(*keypairPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading keypair from %s: %v\n", *keypairPath, err)
		return 1
	}

	pubKeyStr := kp.PublicKeyString()
	fmt.Fprintf(os.Stderr, "Loaded keypair: %s\n", *keypairPath)
	fmt.Fprintf(os.Stderr, "Public key (base64): %s\n", pubKeyStr)

	// Compute a hex-encoded pubkey for Nostr (SHA-256 of the raw Ed25519 public key).
	pubKeyHash := sha256.Sum256(kp.PublicKey)
	nostrPubKey := hex.EncodeToString(pubKeyHash[:])

	// Build the Nostr event content.
	content := fmt.Sprintf("PeerClaw agent identity anchor: %s", pubKeyStr)
	if *agentName != "" {
		content = fmt.Sprintf("PeerClaw agent identity anchor for %s: %s", *agentName, pubKeyStr)
	}

	// Build the unsigned Nostr event (kind 30078 = parameterized replaceable, used for application-specific data).
	tags := [][]string{
		{"d", "peerclaw-identity"},
		{"peerclaw:pubkey", pubKeyStr},
	}
	if *agentName != "" {
		tags = append(tags, []string{"peerclaw:name", *agentName})
	}

	now := time.Now().Unix()
	event := nostrEvent{
		PubKey:    nostrPubKey,
		CreatedAt: now,
		Kind:      30078,
		Tags:      tags,
		Content:   content,
		Sig:       "(sign this event with your Nostr private key before publishing)",
	}

	// Compute event ID as SHA-256 of the serialized event array [0, pubkey, created_at, kind, tags, content].
	serialized, _ := json.Marshal([]any{0, event.PubKey, event.CreatedAt, event.Kind, event.Tags, event.Content})
	eventIDHash := sha256.Sum256(serialized)
	event.ID = hex.EncodeToString(eventIDHash[:])

	fmt.Fprintf(os.Stderr, "\nNostr Identity Anchor Event:\n")
	fmt.Fprintf(os.Stderr, "  Event ID:   %s\n", event.ID)
	fmt.Fprintf(os.Stderr, "  Pubkey:     %s\n", event.PubKey)
	fmt.Fprintf(os.Stderr, "  Kind:       %d\n", event.Kind)
	fmt.Fprintf(os.Stderr, "  Created At: %d\n\n", event.CreatedAt)

	// Output the event JSON to stdout so it can be piped to a Nostr client.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(event)

	fmt.Fprintf(os.Stderr, "\nTo publish this anchor:\n")
	fmt.Fprintf(os.Stderr, "  1. Sign the event with your Nostr private key (replace the 'sig' field)\n")
	fmt.Fprintf(os.Stderr, "  2. Publish to relays: %s\n", *relays)
	fmt.Fprintf(os.Stderr, "  3. Other agents can verify your identity by looking up this event\n")
	return 0
}

func runIdentityVerify(args []string, serverURL string) int {
	fs := flag.NewFlagSet("identity verify", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw identity verify <agent-id>\n")
		return 1
	}

	agentID := fs.Arg(0)
	fmt.Fprintf(os.Stderr, "Initiating endpoint verification for agent %s ...\n", agentID)

	c := client.New(serverURL)
	resp, err := c.VerifyEndpoint(context.Background(), agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Verification result: %s\n", resp.Status)
	if resp.Challenge != "" {
		fmt.Fprintf(os.Stderr, "  Challenge: %s\n", resp.Challenge)
	}
	if resp.Status == "verified" {
		fmt.Println("Agent endpoint verified successfully.")
	} else {
		fmt.Printf("Verification status: %s\n", resp.Status)
	}
	return 0
}
