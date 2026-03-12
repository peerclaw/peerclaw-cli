package acpserve

import (
	"encoding/json"

	coreacp "github.com/peerclaw/peerclaw-core/protocol/acp"
)

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
	AgentID   string        `json:"agent_id"`
	Input     []coreacp.Message `json:"input"`
	SessionID string        `json:"session_id,omitempty"`
	Mode      string        `json:"mode,omitempty"` // sync (default), stream
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

// --- ACP types re-exported from core (H-15) ---

type (
	RunStatus        = coreacp.RunStatus
	Run              = coreacp.Run
	RunError         = coreacp.RunError
	Message          = coreacp.Message
	MessagePart      = coreacp.MessagePart
	AgentManifest    = coreacp.AgentManifest
	ManifestMetadata = coreacp.ManifestMetadata
	CapabilityDef    = coreacp.CapabilityDef
	CreateRunRequest = coreacp.CreateRunRequest
)

const (
	RunStatusCreated    = coreacp.RunStatusCreated
	RunStatusInProgress = coreacp.RunStatusInProgress
	RunStatusAwaiting   = coreacp.RunStatusAwaiting
	RunStatusCompleted  = coreacp.RunStatusCompleted
	RunStatusFailed     = coreacp.RunStatusFailed
	RunStatusCancelling = coreacp.RunStatusCancelling
	RunStatusCancelled  = coreacp.RunStatusCancelled
)
