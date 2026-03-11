package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunACP_NoArgs(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunACP(nil, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("RunACP(nil) returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got:\n%s", out)
	}
}

func TestRunACP_UnknownSubcommand(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunACP([]string{"unknown"}, "http://localhost:8080")

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("RunACP(unknown) returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "unknown acp command") {
		t.Errorf("expected 'unknown acp command', got:\n%s", out)
	}
}
