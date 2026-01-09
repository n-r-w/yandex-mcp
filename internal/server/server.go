// Package server provides MCP server bootstrap and tool registration.
package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MakeHandler creates a typed tool handler from a tool method.
// This is a convenience wrapper that adapts a simple function signature
// to the ToolHandlerFor signature required by mcp.AddTool.
func MakeHandler[In, Out any](
	fn func(context.Context, In) (*Out, error),
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, *Out, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input In) (*mcp.CallToolResult, *Out, error) {
		output, err := fn(ctx, input)
		return nil, output, err
	}
}
