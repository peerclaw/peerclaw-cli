package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConfig_NoSubcommand(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunConfig(nil)

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunConfig() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got:\n%s", out)
	}
}

func TestRunConfig_UnknownSubcommand(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunConfig([]string{"unknown"})

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunConfig() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "unknown config command") {
		t.Errorf("expected 'unknown config command' in output, got:\n%s", out)
	}
}

func TestRunConfig_Show(t *testing.T) {
	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunConfig([]string{"show"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("RunConfig show returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Config file:") {
		t.Errorf("output missing 'Config file:', got:\n%s", out)
	}
	if !strings.Contains(out, "Server:") {
		t.Errorf("output missing 'Server:', got:\n%s", out)
	}
}

func TestLoadCLIConfig_Default(t *testing.T) {
	// When no config file exists, loadCLIConfig should return defaults.
	cfg, err := loadCLIConfig()
	if err != nil {
		t.Fatalf("loadCLIConfig() error = %v", err)
	}
	if cfg.Server == "" {
		t.Error("expected non-empty default server")
	}
	if cfg.Server != defaultServer {
		t.Errorf("Server = %q, want %q", cfg.Server, defaultServer)
	}
}

func TestSaveCLIConfig_RoundTrip(t *testing.T) {
	// Create a temp directory to use as HOME so we don't modify real config.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &CLIConfig{Server: "http://test-server:9090"}
	if err := saveCLIConfig(cfg); err != nil {
		t.Fatalf("saveCLIConfig() error = %v", err)
	}

	// Verify the file was created.
	cfgFile := filepath.Join(tmpDir, ".peerclaw", "config.yaml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", cfgFile)
	}

	// Load and verify.
	loaded, err := loadCLIConfig()
	if err != nil {
		t.Fatalf("loadCLIConfig() error = %v", err)
	}
	if loaded.Server != "http://test-server:9090" {
		t.Errorf("Server = %q, want 'http://test-server:9090'", loaded.Server)
	}
}

func TestRunConfig_Set(t *testing.T) {
	// Create a temp directory to use as HOME.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunConfig([]string{"set", "-server", "http://new-server:8080"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("RunConfig set returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Server set to: http://new-server:8080") {
		t.Errorf("output missing confirmation, got:\n%s", out)
	}

	// Verify the config was actually saved.
	loaded, err := loadCLIConfig()
	if err != nil {
		t.Fatalf("loadCLIConfig() error = %v", err)
	}
	if loaded.Server != "http://new-server:8080" {
		t.Errorf("saved Server = %q, want 'http://new-server:8080'", loaded.Server)
	}
}

func TestRunConfig_SetPositional(t *testing.T) {
	// Create a temp directory to use as HOME.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test positional syntax: peerclaw config set server <url>
	code := RunConfig([]string{"set", "server", "http://positional:9090"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("RunConfig set positional returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Server set to: http://positional:9090") {
		t.Errorf("output missing confirmation, got:\n%s", out)
	}
}

func TestRunConfig_SetMissingArgs(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunConfig([]string{"set"})

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunConfig set returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got:\n%s", out)
	}
}
