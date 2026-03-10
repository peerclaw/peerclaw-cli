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

func TestRunAgent_NoSubcommand(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunAgent(nil, "http://localhost:8080")

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunAgent() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got:\n%s", out)
	}
}

func TestRunAgent_UnknownSubcommand(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunAgent([]string{"unknown"}, "http://localhost:8080")

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunAgent() returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "unknown agent command") {
		t.Errorf("expected 'unknown agent command' in output, got:\n%s", out)
	}
}

func TestRunAgent_ListTable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"Agents": []map[string]any{
				{
					"id":           "agent-1",
					"name":         "TestAgent",
					"status":       "online",
					"protocols":    []string{"a2a"},
					"capabilities": []string{"chat", "search"},
				},
				{
					"id":           "agent-2",
					"name":         "OtherAgent",
					"status":       "offline",
					"protocols":    []string{"mcp", "a2a"},
					"capabilities": []string{"translate"},
				},
			},
			"TotalCount":    2,
			"NextPageToken": "",
		})
	}))
	defer srv.Close()

	// Capture stdout and stderr.
	oldOut := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	oldErr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	outputFormat = "table"
	code := RunAgent([]string{"list"}, srv.URL)

	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	if code != 0 {
		t.Fatalf("RunAgent list returned %d, want 0", code)
	}

	var bufOut bytes.Buffer
	io.Copy(&bufOut, rOut)
	out := bufOut.String()

	var bufErr bytes.Buffer
	io.Copy(&bufErr, rErr)
	errOut := bufErr.String()

	// Check stderr has the total count.
	if !strings.Contains(errOut, "Total: 2 agents") {
		t.Errorf("stderr missing 'Total: 2 agents', got:\n%s", errOut)
	}

	// Check table output contains agent data.
	if !strings.Contains(out, "TestAgent") {
		t.Errorf("output missing 'TestAgent', got:\n%s", out)
	}
	if !strings.Contains(out, "OtherAgent") {
		t.Errorf("output missing 'OtherAgent', got:\n%s", out)
	}
	if !strings.Contains(out, "agent-1") {
		t.Errorf("output missing 'agent-1', got:\n%s", out)
	}
	// Check headers are present.
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("output missing table headers, got:\n%s", out)
	}
}

func TestRunAgent_ListJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"Agents": []map[string]any{
				{
					"id":        "agent-1",
					"name":      "TestAgent",
					"status":    "online",
					"protocols": []string{"a2a"},
				},
			},
			"TotalCount":    1,
			"NextPageToken": "",
		})
	}))
	defer srv.Close()

	oldOut := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	oldErr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr

	outputFormat = "table" // will be overridden by -output flag
	code := RunAgent([]string{"list", "-output", "json"}, srv.URL)

	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	if code != 0 {
		t.Fatalf("RunAgent list -output json returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, rOut)
	out := buf.String()

	// Output should be valid JSON.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput:\n%s", err, out)
	}
	agents, ok := parsed["Agents"].([]any)
	if !ok {
		t.Fatalf("expected Agents array in JSON output")
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
}

func TestRunAgent_ListWithFilters(t *testing.T) {
	var gotProtocol, gotCapability, gotStatus string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotProtocol = r.URL.Query().Get("protocol")
		gotCapability = r.URL.Query().Get("capability")
		gotStatus = r.URL.Query().Get("status")
		json.NewEncoder(w).Encode(map[string]any{
			"Agents":     []any{},
			"TotalCount": 0,
		})
	}))
	defer srv.Close()

	oldOut := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut

	oldErr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr

	outputFormat = "table"
	code := RunAgent([]string{"list", "-protocol", "a2a", "-capability", "chat", "-status", "online"}, srv.URL)

	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	if code != 0 {
		t.Fatalf("RunAgent list with filters returned %d, want 0", code)
	}

	if gotProtocol != "a2a" {
		t.Errorf("protocol filter = %q, want 'a2a'", gotProtocol)
	}
	if gotCapability != "chat" {
		t.Errorf("capability filter = %q, want 'chat'", gotCapability)
	}
	if gotStatus != "online" {
		t.Errorf("status filter = %q, want 'online'", gotStatus)
	}
}

func TestRunAgent_ListConnectionError(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	outputFormat = "table"
	code := RunAgent([]string{"list"}, "http://127.0.0.1:1")

	w.Close()
	os.Stderr = oldErr

	if code != 1 {
		t.Fatalf("RunAgent list returned %d, want 1", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Error:") {
		t.Errorf("expected error output, got:\n%s", out)
	}
}

func TestRunAgent_Help(t *testing.T) {
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := RunAgent([]string{"help"}, "http://localhost:8080")

	w.Close()
	os.Stderr = oldErr

	if code != 0 {
		t.Fatalf("RunAgent help returned %d, want 0", code)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "list") || !strings.Contains(out, "register") {
		t.Errorf("help output missing subcommand info, got:\n%s", out)
	}
}
