package acpserve

import "encoding/json"

// Request is an ndJSON request line on stdin.
type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"` // create_run, get_run, cancel_run, list_agents, get_agent, ping
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is an ndJSON response line on stdout.
type Response struct {
	ID     string `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

// Error describes an error in a Response.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Method params ---

// CreateRunParams are the params for the create_run method.
type CreateRunParams struct {
	AgentID   string    `json:"agent_id"`
	Input     []Message `json:"input"`
	SessionID string    `json:"session_id,omitempty"`
	Mode      string    `json:"mode,omitempty"` // sync (default), stream
}

// GetRunParams are the params for the get_run method.
type GetRunParams struct {
	AgentID string `json:"agent_id"`
	RunID   string `json:"run_id"`
}

// CancelRunParams are the params for the cancel_run method.
type CancelRunParams struct {
	AgentID string `json:"agent_id"`
	RunID   string `json:"run_id"`
}

// ListAgentsParams are the params for the list_agents method.
type ListAgentsParams struct {
	Capability string `json:"capability,omitempty"`
}

// GetAgentParams are the params for the get_agent method.
type GetAgentParams struct {
	AgentID string `json:"agent_id"`
}

// --- ACP types (lightweight copies, JSON-compatible with server/internal/bridge/acp) ---

// RunStatus represents the lifecycle status of an ACP run.
type RunStatus string

const (
	RunStatusCreated    RunStatus = "created"
	RunStatusInProgress RunStatus = "in-progress"
	RunStatusCompleted  RunStatus = "completed"
	RunStatusFailed     RunStatus = "failed"
	RunStatusCancelled  RunStatus = "cancelled"
	RunStatusAwaiting   RunStatus = "awaiting"
	RunStatusCancelling RunStatus = "cancelling"
)

// Run represents an ACP run.
type Run struct {
	AgentName string    `json:"agent_name"`
	RunID     string    `json:"run_id"`
	SessionID string    `json:"session_id,omitempty"`
	Status    RunStatus `json:"status"`
	Input     []Message `json:"input,omitempty"`
	Output    []Message `json:"output,omitempty"`
	Error     *RunError `json:"error,omitempty"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// RunError holds error details for a failed run.
type RunError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

// Message represents an ACP message.
type Message struct {
	Role      string        `json:"role"`
	Parts     []MessagePart `json:"parts"`
	CreatedAt string        `json:"created_at,omitempty"`
}

// MessagePart represents a part of an ACP message.
type MessagePart struct {
	ContentType     string         `json:"content_type"`
	Content         string         `json:"content,omitempty"`
	ContentURL      string         `json:"content_url,omitempty"`
	ContentEncoding string         `json:"content_encoding,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// AgentManifest describes an ACP-compatible agent.
type AgentManifest struct {
	Name               string           `json:"name"`
	Description        string           `json:"description,omitempty"`
	InputContentTypes  []string         `json:"input_content_types,omitempty"`
	OutputContentTypes []string         `json:"output_content_types,omitempty"`
	Metadata           ManifestMetadata `json:"metadata,omitempty"`
}

// ManifestMetadata holds structured metadata for an agent manifest.
type ManifestMetadata struct {
	Capabilities []CapabilityDef `json:"capabilities,omitempty"`
	Domains      []string        `json:"domains,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
}

// CapabilityDef describes a capability.
type CapabilityDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateRunRequest is the request body for POST /acp/{agent_id}/runs.
type CreateRunRequest struct {
	AgentName string    `json:"agent_name"`
	SessionID string    `json:"session_id,omitempty"`
	Input     []Message `json:"input"`
	Mode      string    `json:"mode"`
}
