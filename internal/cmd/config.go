package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CLIConfig holds CLI configuration stored in ~/.peerclaw/config.yaml.
type CLIConfig struct {
	Server string `yaml:"server"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".peerclaw", "config.yaml")
}

func loadCLIConfig() (*CLIConfig, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{Server: defaultServer}, nil
		}
		return nil, err
	}
	var cfg CLIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveCLIConfig(cfg *CLIConfig) error {
	dir := filepath.Dir(configPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

// RunConfig handles the "config" subcommand.
func RunConfig(args []string) int {
	if len(args) < 1 {
		printConfigUsage()
		return 1
	}

	switch args[0] {
	case "show":
		return runConfigShow()
	case "set":
		return runConfigSet(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown config command: %s\n", args[0])
		printConfigUsage()
		return 1
	}
}

func printConfigUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw config <subcommand>

Subcommands:
  show      Show current configuration
  set       Set a configuration value (e.g., peerclaw config set server http://localhost:8080)
`)
}

func runConfigShow() int {
	cfg, err := loadCLIConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Printf("Config file: %s\n", configPath())
	fmt.Printf("Server: %s\n", cfg.Server)
	return 0
}

func runConfigSet(args []string) int {
	fs := flag.NewFlagSet("config set", flag.ExitOnError)
	server := fs.String("server", "", "PeerClaw server URL")
	fs.Parse(args)

	if *server == "" {
		// Try positional args: peerclaw config set server <url>
		if fs.NArg() >= 2 && fs.Arg(0) == "server" {
			*server = fs.Arg(1)
		} else {
			fmt.Fprintf(os.Stderr, "Usage: peerclaw config set -server <url>\n")
			return 1
		}
	}

	cfg, err := loadCLIConfig()
	if err != nil {
		cfg = &CLIConfig{}
	}
	cfg.Server = *server

	if err := saveCLIConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return 1
	}
	fmt.Printf("Server set to: %s\n", cfg.Server)
	return 0
}
