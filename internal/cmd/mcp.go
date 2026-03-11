package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/peerclaw/peerclaw-agent/tools"
	"github.com/peerclaw/peerclaw-core/agentcard"
)

// p2pOnlyTools lists tools that require a full agent (P2P mode) and are
// disabled in API-only mode.
var p2pOnlyTools = []string{
	"send_message", "send_request", "broadcast_message",
	"add_contact", "remove_contact", "list_contacts",
	"get_task", "list_tasks",
}

// RunMCP runs the MCP server command.
func RunMCP(args []string, serverURL string) int {
	if len(args) < 1 {
		printMCPUsage()
		return 1
	}
	switch args[0] {
	case "serve":
		return runMCPServe(args[1:], serverURL)
	default:
		fmt.Fprintf(os.Stderr, "unknown mcp command: %s\n\n", args[0])
		printMCPUsage()
		return 1
	}
}

func runMCPServe(args []string, serverURL string) int {
	fs := flag.NewFlagSet("mcp serve", flag.ContinueOnError)
	var transport string
	var port int
	fs.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	fs.IntVar(&port, "port", 8081, "HTTP port (when transport=http)")
	addServerFlag(fs, &serverURL)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if transport != "stdio" && transport != "http" {
		fmt.Fprintf(os.Stderr, "Error: invalid transport %q (must be stdio or http)\n", transport)
		return 1
	}

	// Create API client for server communication.
	apiClient := tools.NewAPIClient(serverURL)

	// Create handler in API-only mode (P2P tools disabled).
	handler := tools.NewHandler(tools.Options{
		APIClient: apiClient,
		Disabled:  p2pOnlyTools,
	})

	// Create MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "peerclaw",
		Version: "0.5.0",
	}, nil)

	// Register available tools.
	for _, tool := range handler.AvailableTools() {
		registerMCPTool(server, handler, tool)
	}

	// Register resources.
	registerDirectoryResource(server, apiClient)

	// Run server.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if transport == "stdio" {
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			if ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
		}
	} else {
		httpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return server
		}, nil)
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: httpHandler,
		}

		go func() {
			<-ctx.Done()
			srv.Close()
		}()

		fmt.Fprintf(os.Stderr, "MCP server listening on :%d\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	return 0
}

// registerMCPTool bridges an agentcard.Tool to the MCP server using the raw
// ToolHandler API so we can pass our pre-defined JSON Schemas directly.
func registerMCPTool(server *mcp.Server, handler *tools.Handler, tool agentcard.Tool) {
	toolName := tool.Name // capture for closure

	// Convert json.RawMessage schema to map[string]any for MCP SDK.
	var inputSchema any
	if len(tool.InputSchema) > 0 {
		var m map[string]any
		if err := json.Unmarshal(tool.InputSchema, &m); err == nil {
			inputSchema = m
		}
	}

	mcpTool := &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: inputSchema,
		Annotations: toolAnnotations(tool.Name),
	}

	server.AddTool(mcpTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// req.Params.Arguments is json.RawMessage (from CallToolParamsRaw).
		input := req.Params.Arguments
		if input == nil {
			input = json.RawMessage("{}")
		}

		result, err := handler.Handle(ctx, toolName, input)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(result)}},
		}, nil
	})
}

// toolAnnotations returns MCP tool annotations for the given tool name.
func toolAnnotations(name string) *mcp.ToolAnnotations {
	switch name {
	case "discover_agents", "get_agent_profile", "check_reputation":
		return &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: ptrBool(false),
		}
	case "invoke_agent":
		return &mcp.ToolAnnotations{
			DestructiveHint: ptrBool(false),
		}
	default:
		return nil
	}
}

func ptrBool(b bool) *bool { return &b }

// registerDirectoryResource exposes the agent directory as an MCP resource.
func registerDirectoryResource(server *mcp.Server, apiClient *tools.APIClient) {
	server.AddResource(&mcp.Resource{
		URI:         "peerclaw://directory",
		Name:        "Agent Directory",
		Description: "Browse the PeerClaw agent directory",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := apiClient.BrowseDirectory(ctx, tools.DirectoryRequest{PageSize: 50})
		if err != nil {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     fmt.Sprintf(`{"error": %q}`, err.Error()),
				}},
			}, nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}

func printMCPUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw mcp <command> [options]

Commands:
  serve     Start MCP server for AI tool integration

Options (serve):
  --transport   Transport type: stdio (default) or http
  --port        HTTP port when transport=http (default: 8081)
  --server      PeerClaw server URL (default: %s)

Examples:
  peerclaw mcp serve
  peerclaw mcp serve --server http://localhost:8080
  peerclaw mcp serve --transport http --port 8081
`, defaultServer)
}
