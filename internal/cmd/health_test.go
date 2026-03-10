package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunHealth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"components": map[string]string{
				"database": "healthy",
				"bridge":   "healthy",
			},
			"connected_agents":  3,
			"registered_agents": 5,
		})
	}))
	defer srv.Close()

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset outputFormat.
	outputFormat = "table"
	code := RunHealth(nil, srv.URL)

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("RunHealth() returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Status: ok") {
		t.Errorf("output missing 'Status: ok', got:\n%s", out)
	}
	if !strings.Contains(out, "Connected Agents: 3") {
		t.Errorf("output missing 'Connected Agents: 3', got:\n%s", out)
	}
	if !strings.Contains(out, "Registered Agents: 5") {
		t.Errorf("output missing 'Registered Agents: 5', got:\n%s", out)
	}
	if !strings.Contains(out, "database: healthy") {
		t.Errorf("output missing 'database: healthy', got:\n%s", out)
	}
}

func TestRunHealth_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":            "ok",
			"connected_agents":  1,
			"registered_agents": 2,
		})
	}))
	defer srv.Close()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputFormat = "table" // will be overridden by -output flag
	code := RunHealth([]string{"-output", "json"}, srv.URL)

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("RunHealth() returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	// The output should be valid JSON.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput:\n%s", err, out)
	}
	if parsed["status"] != "ok" {
		t.Errorf("JSON status = %v, want 'ok'", parsed["status"])
	}
}

func TestRunHealth_ConnectionError(t *testing.T) {
	// Capture stderr.
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	outputFormat = "table"
	// Use a server URL that will refuse connections.
	code := RunHealth(nil, "http://127.0.0.1:1")

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunHealth() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Error:") {
		t.Errorf("expected error output, got:\n%s", out)
	}
}

func TestRunHealth_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer srv.Close()

	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	outputFormat = "table"
	code := RunHealth(nil, srv.URL)

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunHealth() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Error:") {
		t.Errorf("expected error output, got:\n%s", out)
	}
}
