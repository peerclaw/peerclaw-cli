package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// serviceManager abstracts OS-level service operations.
type serviceManager interface {
	Install(cfg serviceConfig) error
	Uninstall() error
	Status() (string, error)
	Logs(lines int, follow bool) error
}

// serviceConfig holds configuration for the managed service.
type serviceConfig struct {
	BinaryPath string // absolute path to the peerclaw binary
	ServerURL  string // PeerClaw server URL
	AgentID    string // agent identity (required for heartbeat mode)
	Mode       string // "mcp" or "heartbeat"
	Force      bool   // overwrite existing service
}

// execArgs returns the command-line arguments for the service process.
func (c serviceConfig) execArgs() []string {
	switch c.Mode {
	case "heartbeat":
		return []string{c.BinaryPath, "agent", "heartbeat", c.AgentID, "--loop", "--server", c.ServerURL}
	default: // "mcp"
		return []string{c.BinaryPath, "mcp", "serve", "--server", c.ServerURL}
	}
}

// RunService handles the "service" subcommand.
func RunService(args []string, serverURL string) int {
	if len(args) < 1 {
		printServiceUsage()
		return 1
	}

	mgr := newServiceManager()

	switch args[0] {
	case "install":
		return runServiceInstall(args[1:], serverURL, mgr)
	case "uninstall":
		return runServiceUninstall(mgr)
	case "status":
		return runServiceStatus(mgr)
	case "logs":
		return runServiceLogs(args[1:], mgr)
	default:
		fmt.Fprintf(os.Stderr, "unknown service command: %s\n\n", args[0])
		printServiceUsage()
		return 1
	}
}

func runServiceInstall(args []string, serverURL string, mgr serviceManager) int {
	fs := flag.NewFlagSet("service install", flag.ContinueOnError)
	mode := fs.String("mode", "mcp", "Service mode: mcp (default) or heartbeat")
	force := fs.Bool("force", false, "Overwrite existing service file")
	addServerFlag(fs, &serverURL)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *mode != "mcp" && *mode != "heartbeat" {
		fmt.Fprintf(os.Stderr, "Error: invalid mode %q (must be mcp or heartbeat)\n", *mode)
		return 1
	}

	// Resolve binary path.
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine binary path: %v\n", err)
		return 1
	}
	binPath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve binary path: %v\n", err)
		return 1
	}

	// Warn if binary is in a temporary location.
	if strings.Contains(binPath, "/tmp/") || strings.Contains(binPath, "go-build") {
		fmt.Fprintf(os.Stderr, "Warning: binary path %q looks temporary — the service may break after cleanup\n", binPath)
	}

	// Load config for agent_id.
	cfg, _ := loadCLIConfig()
	agentID := ""
	if cfg != nil {
		agentID = cfg.AgentID
	}

	if *mode == "heartbeat" && agentID == "" {
		fmt.Fprintln(os.Stderr, "Error: heartbeat mode requires agent_id in config")
		fmt.Fprintln(os.Stderr, "Run: peerclaw agent claim <name> --server <url>")
		return 1
	}

	svcCfg := serviceConfig{
		BinaryPath: binPath,
		ServerURL:  serverURL,
		AgentID:    agentID,
		Mode:       *mode,
		Force:      *force,
	}

	if err := mgr.Install(svcCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("Service installed and started successfully.")
	fmt.Printf("  Mode:   %s\n", svcCfg.Mode)
	fmt.Printf("  Binary: %s\n", svcCfg.BinaryPath)
	fmt.Printf("  Server: %s\n", svcCfg.ServerURL)
	if svcCfg.AgentID != "" {
		fmt.Printf("  Agent:  %s\n", svcCfg.AgentID)
	}
	fmt.Println()
	fmt.Println("Manage with:")
	fmt.Println("  peerclaw service status")
	fmt.Println("  peerclaw service logs -f")
	fmt.Println("  peerclaw service uninstall")
	return 0
}

func runServiceUninstall(mgr serviceManager) int {
	if err := mgr.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Println("Service uninstalled.")
	return 0
}

func runServiceStatus(mgr serviceManager) int {
	status, err := mgr.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(status)
	return 0
}

func runServiceLogs(args []string, mgr serviceManager) int {
	fs := flag.NewFlagSet("service logs", flag.ContinueOnError)
	lines := fs.Int("n", 50, "Number of log lines to show")
	follow := fs.Bool("f", false, "Follow log output")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if err := mgr.Logs(*lines, *follow); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func printServiceUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw service <command> [options]

Commands:
  install     Install and start the agent as an OS service
  uninstall   Stop and remove the agent service
  status      Show service status
  logs        View service logs

Options (install):
  --mode      Service mode: mcp (default) or heartbeat
  --force     Overwrite existing service file
  --server    PeerClaw server URL

Options (logs):
  -n          Number of lines to show (default: 50)
  -f          Follow log output

Examples:
  peerclaw service install
  peerclaw service install --mode heartbeat
  peerclaw service install --force
  peerclaw service status
  peerclaw service logs -n 20 -f
  peerclaw service uninstall
`)
}
