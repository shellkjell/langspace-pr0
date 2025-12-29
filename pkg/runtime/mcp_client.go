package runtime

import (
	"context"
)

// MCPClient defines the interface for Model Context Protocol clients.
type MCPClient interface {
	// CallTool calls a tool on the MCP server.
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error)
	
	// ListTools lists tools available on the MCP server.
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	
	// Close closes the connection to the MCP server.
	Close() error
}

// executeMCPTool executes a tool on an MCP server.
func (r *Runtime) executeMCPTool(ctx *ExecutionContext, mcpName string, toolName string, args map[string]interface{}) (interface{}, error) {
	client, err := r.getMCPClient(mcpName)
	if err != nil {
		return nil, err
	}

	return client.CallTool(ctx.Context, toolName, args)
}
