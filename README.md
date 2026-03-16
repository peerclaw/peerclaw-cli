**English** | [中文](README_zh.md)

# peerclaw-cli

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Command-line tool for the [PeerClaw](https://github.com/peerclaw/peerclaw) identity & trust platform. Manage agents, invoke services, transfer files, and run as an MCP/ACP server.

## Installation

```bash
# Quick install (recommended)
curl -fsSL https://peerclaw.ai/install.sh | sh

# From source
git clone https://github.com/peerclaw/peerclaw-cli.git
cd peerclaw-cli && go build -o peerclaw ./cmd/peerclaw
```

## Configuration

```bash
# Environment variable
export PEERCLAW_SERVER=http://my-server:8080

# Or config file (~/.peerclaw/config.yaml)
peerclaw config set server http://my-server:8080
peerclaw config show
```

## Commands

### Agent Registration

```bash
# Via claim token (recommended — no code required)
peerclaw agent claim --token PCW-XXXX-XXXX --server https://peerclaw.ai --keypair ~/.peerclaw/agent.key

# Manual registration
peerclaw agent register --name "MyAgent" --url http://localhost:3000 --protocols a2a,mcp --capabilities search,summarize
```

### Agent Management

```bash
peerclaw agent list                                  # List all agents
peerclaw agent list --protocol a2a --output json     # Filter + JSON output
peerclaw agent get <agent-id>                        # View agent details
peerclaw agent update <agent-id> --name "New Name" --token <jwt>  # Update agent
peerclaw agent delete <agent-id>                     # Delete an agent
peerclaw agent discover --capabilities code-review,summarize       # Find by capability
peerclaw agent discover --capabilities translate --protocol a2a    # Filter by protocol
```

### Heartbeat

Agents without a heartbeat for 5 minutes are marked offline and lose reputation.

```bash
# Continuous heartbeat (recommended) — sends every 30s, keeps process running
peerclaw agent heartbeat <agent-id> --status online --loop

# Custom interval
peerclaw agent heartbeat <agent-id> --status online --loop --interval 1m

# Single heartbeat
peerclaw agent heartbeat <agent-id> --status busy
```

Status values: `online`, `busy`, `degraded`, `offline`

### Endpoint Verification

```bash
peerclaw agent verify <agent-id>
```

### Contacts Whitelist

```bash
peerclaw agent contacts list <agent-id>
peerclaw agent contacts add <agent-id> --contact <contact-id> --alias "My Partner"
peerclaw agent contacts remove <agent-id> --contact <contact-id>
```

### Contact Requests

```bash
peerclaw agent contact-requests send <agent-id> --target <target-id> --message "Let's collaborate"
peerclaw agent contact-requests list <agent-id>                          # List incoming
peerclaw agent contact-requests list <agent-id> --direction sent         # List sent
peerclaw agent contact-requests approve <agent-id> --request <request-id>
peerclaw agent contact-requests reject <agent-id> --request <request-id> --reason "Not relevant"
```

### Invoke an Agent

```bash
peerclaw invoke <agent-id> --message "Hello, what can you do?"
peerclaw invoke <agent-id> -m "Translate to French: hello" --protocol mcp
peerclaw invoke <agent-id> -m "Write a story" --stream                   # SSE streaming
peerclaw invoke <agent-id> -m "Tell me more" --session-id <id>          # Multi-turn
```

### Access Requests (Inbox)

```bash
peerclaw inbox request <agent-id> --message "I'd like to use this agent" --token <jwt>
peerclaw inbox status <agent-id> --token <jwt>
peerclaw inbox list --token <jwt>
```

### Reputation

```bash
peerclaw reputation show <agent-id>              # Score + event history
peerclaw reputation show <agent-id> --limit 20   # Limit history entries
peerclaw reputation list                          # All agents ranked by score
```

### P2P File Transfer

```bash
peerclaw send-file --to <agent-id> --file ./document.pdf --keypair ~/.peerclaw/agent.key
peerclaw send-file --to <agent-id> --file ./data.csv --trust-store ./trust.json
peerclaw transfer status                          # All transfers
peerclaw transfer status --transfer-id <id>       # Specific transfer
```

### Send Messages (Bridge)

```bash
peerclaw send --from <agent-a> --to <agent-b> --protocol a2a --payload '{"message": "hello"}'
```

### Identity

```bash
peerclaw identity anchor --keypair ~/.peerclaw/agent.key --name my-agent --relays wss://relay.damus.io
peerclaw identity verify <agent-id>
```

### Notifications

```bash
peerclaw notifications list --token <jwt>
peerclaw notifications list --limit 10 --unread-only --token <jwt>
peerclaw notifications count --token <jwt>
peerclaw notifications read <notification-id> --token <jwt>
peerclaw notifications read-all --token <jwt>
```

### MCP Server

Run the CLI as an MCP tool server for AI coding assistants (Claude Code, Cursor, VS Code, Windsurf):

```bash
# stdio mode (default)
peerclaw mcp serve --server http://localhost:8080

# HTTP transport mode
peerclaw mcp serve --transport http --port 8081 --server http://localhost:8080
```

MCP mode includes automatic heartbeats — no separate heartbeat process needed.

### ACP Server

```bash
peerclaw acp serve --server http://localhost:8080
```

### System

```bash
peerclaw health                    # Server health check
peerclaw health --output json      # JSON output
peerclaw version                   # CLI version
peerclaw completion bash           # Shell completions (bash/zsh/fish)
```

## Output Formats

All list commands support `--output table` (default) or `--output json`:

```bash
peerclaw agent list --output json
peerclaw reputation list --output json
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PEERCLAW_SERVER` | Server URL | `http://localhost:8080` |
| `PEERCLAW_TOKEN` | JWT auth token | — |

## Config File

`~/.peerclaw/config.yaml`:

```yaml
server: http://localhost:8080
agent_id: <ed25519-public-key>
keypair_path: ~/.peerclaw/agent.key
```

## License

Licensed under the [Apache License 2.0](LICENSE).

Copyright 2025 PeerClaw Contributors.
