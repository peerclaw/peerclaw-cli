package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunMCP_NoArgs(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunMCP(nil, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("RunMCP(nil) returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got:\n%s", out)
	}
}

func TestRunMCP_UnknownSubcommand(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunMCP([]string{"unknown"}, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("RunMCP(unknown) returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "unknown mcp command") {
		t.Errorf("expected 'unknown mcp command', got:\n%s", out)
	}
}

func TestRunMCP_InvalidTransport(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunMCP([]string{"serve", "--transport", "grpc"}, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("RunMCP(--transport grpc) returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "invalid transport") {
		t.Errorf("expected 'invalid transport' error, got:\n%s", out)
	}
}

func TestRunMCP_DefaultFlags(t *testing.T) {
	// Verify flag parsing works with default values by checking
	// that valid flags don't cause a parse error.
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// This will start the server but exit quickly because stdin closes.
	// We only care that flag parsing succeeds (doesn't return 1 from parse error).
	code := RunMCP([]string{"serve", "--server", "http://example.com:8080", "--transport", "stdio"}, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	// The server will either exit cleanly (0) or with a transport error,
	// but should NOT fail with "invalid transport" or flag parse errors.
	if strings.Contains(out, "invalid transport") {
		t.Errorf("unexpected 'invalid transport' error with valid flags")
	}
	_ = code // exit code depends on stdin availability
}
