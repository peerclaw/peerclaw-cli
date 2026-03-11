package acpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// helper: send ndJSON lines to a Server and return the output lines.
func runLines(t *testing.T, handler http.Handler, lines ...string) []Response {
	t.Helper()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	input := strings.Join(lines, "\n") + "\n"
	var out bytes.Buffer
	srv := New(ts.URL, strings.NewReader(input), &out)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var responses []Response
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("unmarshal response %q: %v", line, err)
		}
		responses = append(responses, resp)
	}
	return responses
}

func TestServer_Ping(t *testing.T) {
	resps := runLines(t, http.NewServeMux(), `{"id":"1","method":"ping"}`)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	r := resps[0]
	if r.ID != "1" {
		t.Errorf("id = %q, want 1", r.ID)
	}
	if r.Error != nil {
		t.Fatalf("unexpected error: %+v", r.Error)
	}
	// Result should contain {"status":"ok"}.
	data, _ := json.Marshal(r.Result)
	if !strings.Contains(string(data), `"status":"ok"`) {
		t.Errorf("result = %s, want status ok", string(data))
	}
}

func TestServer_InvalidJSON(t *testing.T) {
	resps := runLines(t, http.NewServeMux(), `{not valid json}`)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if resps[0].Error.Code != "parse_error" {
		t.Errorf("error code = %q, want parse_error", resps[0].Error.Code)
	}
}

func TestServer_UnknownMethod(t *testing.T) {
	resps := runLines(t, http.NewServeMux(), `{"id":"2","method":"do_magic"}`)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resps[0].Error.Code != "unknown_method" {
		t.Errorf("error code = %q, want unknown_method", resps[0].Error.Code)
	}
}

func TestServer_CreateRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /acp/{agent_id}/runs", func(w http.ResponseWriter, r *http.Request) {
		agentID := r.PathValue("agent_id")
		var req CreateRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		run := Run{
			AgentName: agentID,
			RunID:     "run-001",
			SessionID: "sess-001",
			Status:    RunStatusCreated,
			Input:     req.Input,
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(run)
	})

	params, _ := json.Marshal(CreateRunParams{
		AgentID: "agent-abc",
		Input: []Message{{
			Role:  "user",
			Parts: []MessagePart{{ContentType: "text/plain", Content: "hello"}},
		}},
	})
	line := fmt.Sprintf(`{"id":"3","method":"create_run","params":%s}`, string(params))

	resps := runLines(t, mux, line)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	r := resps[0]
	if r.ID != "3" {
		t.Errorf("id = %q, want 3", r.ID)
	}
	if r.Error != nil {
		t.Fatalf("unexpected error: %+v", r.Error)
	}

	// Decode result as Run.
	data, _ := json.Marshal(r.Result)
	var run Run
	if err := json.Unmarshal(data, &run); err != nil {
		t.Fatalf("decode run: %v", err)
	}
	if run.RunID != "run-001" {
		t.Errorf("run_id = %q, want run-001", run.RunID)
	}
	if run.Status != RunStatusCreated {
		t.Errorf("status = %q, want created", run.Status)
	}
}

func TestServer_GetRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /acp/{agent_id}/runs/{run_id}", func(w http.ResponseWriter, r *http.Request) {
		run := Run{
			AgentName: r.PathValue("agent_id"),
			RunID:     r.PathValue("run_id"),
			Status:    RunStatusCompleted,
			Output: []Message{{
				Role:  "agent",
				Parts: []MessagePart{{ContentType: "text/plain", Content: "done"}},
			}},
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:01Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	})

	params, _ := json.Marshal(GetRunParams{AgentID: "agent-abc", RunID: "run-002"})
	line := fmt.Sprintf(`{"id":"4","method":"get_run","params":%s}`, string(params))

	resps := runLines(t, mux, line)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error != nil {
		t.Fatalf("unexpected error: %+v", resps[0].Error)
	}

	data, _ := json.Marshal(resps[0].Result)
	var run Run
	json.Unmarshal(data, &run)
	if run.RunID != "run-002" {
		t.Errorf("run_id = %q, want run-002", run.RunID)
	}
	if run.Status != RunStatusCompleted {
		t.Errorf("status = %q, want completed", run.Status)
	}
}

func TestServer_GetRun_Cached(t *testing.T) {
	// First create a run, then get it without hitting the server.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /acp/{agent_id}/runs", func(w http.ResponseWriter, r *http.Request) {
		run := Run{
			AgentName: "agent-abc",
			RunID:     "run-cached",
			Status:    RunStatusCreated,
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(run)
	})
	// No GET handler — if get_run hits the server, it would 404.

	createParams, _ := json.Marshal(CreateRunParams{
		AgentID: "agent-abc",
		Input: []Message{{
			Role:  "user",
			Parts: []MessagePart{{ContentType: "text/plain", Content: "hi"}},
		}},
	})
	getParams, _ := json.Marshal(GetRunParams{RunID: "run-cached"})

	lines := []string{
		fmt.Sprintf(`{"id":"c1","method":"create_run","params":%s}`, string(createParams)),
		fmt.Sprintf(`{"id":"c2","method":"get_run","params":%s}`, string(getParams)),
	}

	resps := runLines(t, mux, lines...)
	if len(resps) != 2 {
		t.Fatalf("got %d responses, want 2", len(resps))
	}
	if resps[1].Error != nil {
		t.Fatalf("get_run from cache failed: %+v", resps[1].Error)
	}

	data, _ := json.Marshal(resps[1].Result)
	var run Run
	json.Unmarshal(data, &run)
	if run.RunID != "run-cached" {
		t.Errorf("run_id = %q, want run-cached", run.RunID)
	}
}

func TestServer_CancelRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /acp/{agent_id}/runs/{run_id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		run := Run{
			AgentName: r.PathValue("agent_id"),
			RunID:     r.PathValue("run_id"),
			Status:    RunStatusCancelled,
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:02Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	})

	params, _ := json.Marshal(CancelRunParams{AgentID: "agent-abc", RunID: "run-003"})
	line := fmt.Sprintf(`{"id":"5","method":"cancel_run","params":%s}`, string(params))

	resps := runLines(t, mux, line)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error != nil {
		t.Fatalf("unexpected error: %+v", resps[0].Error)
	}

	data, _ := json.Marshal(resps[0].Result)
	var run Run
	json.Unmarshal(data, &run)
	if run.Status != RunStatusCancelled {
		t.Errorf("status = %q, want cancelled", run.Status)
	}
}

func TestServer_ListAgents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/directory", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"agents": []map[string]any{
				{"id": "a1", "name": "Alpha", "description": "Agent Alpha", "capabilities": []string{"chat"}},
				{"id": "a2", "name": "Beta", "description": "Agent Beta", "capabilities": []string{"code"}},
			},
		})
	})

	resps := runLines(t, mux, `{"id":"6","method":"list_agents"}`)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error != nil {
		t.Fatalf("unexpected error: %+v", resps[0].Error)
	}

	data, _ := json.Marshal(resps[0].Result)
	var result struct {
		Agents []AgentManifest `json:"agents"`
	}
	json.Unmarshal(data, &result)
	if len(result.Agents) != 2 {
		t.Fatalf("got %d agents, want 2", len(result.Agents))
	}
	if result.Agents[0].Name != "Alpha" {
		t.Errorf("agent[0].name = %q, want Alpha", result.Agents[0].Name)
	}
}

func TestServer_GetAgent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /acp/{agent_id}/agents", func(w http.ResponseWriter, r *http.Request) {
		manifest := AgentManifest{
			Name:               r.PathValue("agent_id"),
			Description:        "Test Agent",
			InputContentTypes:  []string{"text/plain"},
			OutputContentTypes: []string{"text/plain"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manifest)
	})

	params, _ := json.Marshal(GetAgentParams{AgentID: "agent-xyz"})
	line := fmt.Sprintf(`{"id":"7","method":"get_agent","params":%s}`, string(params))

	resps := runLines(t, mux, line)
	if len(resps) != 1 {
		t.Fatalf("got %d responses, want 1", len(resps))
	}
	if resps[0].Error != nil {
		t.Fatalf("unexpected error: %+v", resps[0].Error)
	}

	data, _ := json.Marshal(resps[0].Result)
	var manifest AgentManifest
	json.Unmarshal(data, &manifest)
	if manifest.Name != "agent-xyz" {
		t.Errorf("name = %q, want agent-xyz", manifest.Name)
	}
	if manifest.Description != "Test Agent" {
		t.Errorf("description = %q, want Test Agent", manifest.Description)
	}
}

func TestServer_BlankLines(t *testing.T) {
	// Blank lines should be silently skipped.
	resps := runLines(t, http.NewServeMux(),
		"",
		`{"id":"b1","method":"ping"}`,
		"   ",
		`{"id":"b2","method":"ping"}`,
	)
	if len(resps) != 2 {
		t.Fatalf("got %d responses, want 2", len(resps))
	}
}
