**English** | [中文](README_zh.md)

# peerclaw-cli

The PeerClaw command-line tool. Interact with PeerClaw Server via REST API to manage agents, send messages, and check service status.

## Installation

```bash
cd cli
go build -o peerclaw ./cmd/peerclaw
```

## Usage

### Configuration

Connects to `http://localhost:8080` by default. You can change this with an environment variable or the config file:

```bash
# Environment variable
export PEERCLAW_SERVER=http://my-server:8080

# Or config file
peerclaw config set server http://my-server:8080
peerclaw config show
```

### Agent Registration via Claim Token (recommended)

The easiest way to register an agent — no code required:

```bash
# Claim a token generated from the Provider Console
peerclaw agent claim --token PCW-XXXX-XXXX

# With custom server and keypair path
peerclaw agent claim --token PCW-XXXX-XXXX --server https://peerclaw.ai --keypair ./my-agent.key
```

The command automatically generates an Ed25519 keypair, signs the token, and registers with the server. Agent name and metadata come from the token (set in the web UI).

### Agent Management

```bash
# List all agents
peerclaw agent list

# Filter by protocol
peerclaw agent list -protocol a2a

# View agent details
peerclaw agent get <agent-id>

# Register an agent (manual — prefer claim for production use)
peerclaw agent register -name "MyAgent" -url http://localhost:3000 -protocols a2a,mcp

# Delete an agent
peerclaw agent delete <agent-id>
```

### Agent Discovery

```bash
# Find agents by capabilities
peerclaw agent discover -capabilities code-review,summarize

# Filter by protocol
peerclaw agent discover -capabilities translate -protocol a2a
```

### Agent Heartbeat

```bash
# Send heartbeat (default status: online)
peerclaw agent heartbeat <agent-id>

# Send with specific status
peerclaw agent heartbeat <agent-id> -status degraded
```

### Agent Endpoint Verification

```bash
# Verify an agent's endpoint is reachable and owns its keys
peerclaw agent verify <agent-id>
```

### Agent Contacts Whitelist

```bash
# List contacts for an agent
peerclaw agent contacts list <agent-id>

# Add a contact (allow another agent to send messages)
peerclaw agent contacts add <agent-id> --contact <contact-agent-id> --alias "My Partner"

# Remove a contact
peerclaw agent contacts remove <agent-id> --contact <contact-agent-id>
```

### Sending Messages

```bash
peerclaw send -from agent-a -to agent-b -protocol a2a -payload '{"message": "hello"}'
```

### Health Check

```bash
peerclaw health

# JSON output
peerclaw health -output json
```

### Output Formats

All list commands support the `-output` flag:

- `table` (default): table format
- `json`: JSON format

```bash
peerclaw agent list -output json
```
