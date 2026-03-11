package acpserve

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Server implements an ACP stdio bridge that reads ndJSON requests from
// a reader (stdin) and writes ndJSON responses to a writer (stdout),
// proxying calls to the PeerClaw server's ACP HTTP bridge.
type Server struct {
	baseURL    string
	httpClient *http.Client
	runs       sync.Map // runID → *Run
	reader     *bufio.Scanner
	writer     io.Writer
	mu         sync.Mutex // serialise writes
}

// New creates a new ACP stdio bridge server.
func New(baseURL string, reader io.Reader, writer io.Writer) *Server {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1 MB line limit
	return &Server{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		reader: scanner,
		writer: writer,
	}
}

// Run reads ndJSON lines from stdin and dispatches each request.
// It blocks until the reader is exhausted or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	for s.reader.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := s.reader.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeResponse(Response{
				Error: &Error{Code: "parse_error", Message: "invalid JSON: " + err.Error()},
			})
			continue
		}
		s.dispatch(ctx, &req)
	}
	if err := s.reader.Err(); err != nil && ctx.Err() == nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	return nil
}

func (s *Server) dispatch(ctx context.Context, req *Request) {
	switch req.Method {
	case "ping":
		s.handlePing(req)
	case "create_run":
		s.handleCreateRun(ctx, req)
	case "get_run":
		s.handleGetRun(ctx, req)
	case "cancel_run":
		s.handleCancelRun(ctx, req)
	case "list_agents":
		s.handleListAgents(ctx, req)
	case "get_agent":
		s.handleGetAgent(ctx, req)
	default:
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "unknown_method", Message: "unknown method: " + req.Method},
		})
	}
}

func (s *Server) handlePing(req *Request) {
	s.writeResponse(Response{
		ID:     req.ID,
		Result: map[string]string{"status": "ok"},
	})
}

func (s *Server) handleCreateRun(ctx context.Context, req *Request) {
	var params CreateRunParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "invalid params: " + err.Error()},
		})
		return
	}
	if params.AgentID == "" {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "agent_id is required"},
		})
		return
	}

	mode := params.Mode
	if mode == "" {
		mode = "sync"
	}

	createReq := CreateRunRequest{
		AgentName: params.AgentID,
		SessionID: params.SessionID,
		Input:     params.Input,
		Mode:      mode,
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "marshal request: " + err.Error()},
		})
		return
	}

	url := s.baseURL + "/acp/" + params.AgentID + "/runs"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: err.Error()},
		})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: "create run: " + err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(errBody))},
		})
		return
	}

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "decode run: " + err.Error()},
		})
		return
	}

	s.runs.Store(run.RunID, &run)
	s.writeResponse(Response{ID: req.ID, Result: &run})
}

func (s *Server) handleGetRun(ctx context.Context, req *Request) {
	var params GetRunParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "invalid params: " + err.Error()},
		})
		return
	}
	if params.RunID == "" {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "run_id is required"},
		})
		return
	}

	// Check local cache first.
	if v, ok := s.runs.Load(params.RunID); ok {
		s.writeResponse(Response{ID: req.ID, Result: v})
		return
	}

	// Fall back to server.
	if params.AgentID == "" {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "agent_id is required for uncached runs"},
		})
		return
	}

	url := s.baseURL + "/acp/" + params.AgentID + "/runs/" + params.RunID
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: err.Error()},
		})
		return
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: "get run: " + err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(errBody))},
		})
		return
	}

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "decode run: " + err.Error()},
		})
		return
	}

	s.runs.Store(run.RunID, &run)
	s.writeResponse(Response{ID: req.ID, Result: &run})
}

func (s *Server) handleCancelRun(ctx context.Context, req *Request) {
	var params CancelRunParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "invalid params: " + err.Error()},
		})
		return
	}
	if params.AgentID == "" || params.RunID == "" {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "agent_id and run_id are required"},
		})
		return
	}

	url := s.baseURL + "/acp/" + params.AgentID + "/runs/" + params.RunID + "/cancel"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: err.Error()},
		})
		return
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: "cancel run: " + err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(errBody))},
		})
		return
	}

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "decode run: " + err.Error()},
		})
		return
	}

	s.runs.Store(run.RunID, &run)
	s.writeResponse(Response{ID: req.ID, Result: &run})
}

func (s *Server) handleListAgents(ctx context.Context, req *Request) {
	var params ListAgentsParams
	if req.Params != nil {
		_ = json.Unmarshal(req.Params, &params)
	}

	url := s.baseURL + "/api/v1/directory"
	if params.Capability != "" {
		url += "?capability=" + params.Capability
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: err.Error()},
		})
		return
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: "list agents: " + err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(errBody))},
		})
		return
	}

	// The directory API returns {"agents": [...]} with AgentProfile objects.
	// Convert to ACP AgentManifest format.
	var dirResp struct {
		Agents []struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Description  string   `json:"description"`
			Capabilities []string `json:"capabilities"`
		} `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dirResp); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "decode directory: " + err.Error()},
		})
		return
	}

	manifests := make([]AgentManifest, 0, len(dirResp.Agents))
	for _, a := range dirResp.Agents {
		name := a.Name
		if name == "" {
			name = a.ID
		}
		caps := make([]CapabilityDef, 0, len(a.Capabilities))
		for _, c := range a.Capabilities {
			caps = append(caps, CapabilityDef{Name: c})
		}
		manifests = append(manifests, AgentManifest{
			Name:               name,
			Description:        a.Description,
			InputContentTypes:  []string{"text/plain", "application/json"},
			OutputContentTypes: []string{"text/plain", "application/json"},
			Metadata:           ManifestMetadata{Capabilities: caps},
		})
	}

	s.writeResponse(Response{ID: req.ID, Result: map[string]any{"agents": manifests}})
}

func (s *Server) handleGetAgent(ctx context.Context, req *Request) {
	var params GetAgentParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "invalid params: " + err.Error()},
		})
		return
	}
	if params.AgentID == "" {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "invalid_params", Message: "agent_id is required"},
		})
		return
	}

	url := s.baseURL + "/acp/" + params.AgentID + "/agents"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: err.Error()},
		})
		return
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: "get agent: " + err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "server_error", Message: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(errBody))},
		})
		return
	}

	var manifest AgentManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		s.writeResponse(Response{
			ID:    req.ID,
			Error: &Error{Code: "internal", Message: "decode manifest: " + err.Error()},
		})
		return
	}

	s.writeResponse(Response{ID: req.ID, Result: &manifest})
}

// writeResponse serialises a Response as a single ndJSON line.
func (s *Server) writeResponse(resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		// Last resort: write a bare error.
		data = []byte(`{"error":{"code":"internal","message":"marshal response failed"}}`)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.writer.Write(append(data, '\n'))
}
