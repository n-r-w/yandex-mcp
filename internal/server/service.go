package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server encapsulates an MCP server instance.
type Server struct {
	mcpServer *mcp.Server
}

// New initializes an MCP server with the given registrators.
func New(serverVersion string, registrators []IToolsRegistrator) (*Server, error) {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{ //nolint:exhaustruct // optional fields use defaults
			Name:    serverName,
			Version: serverVersion,
			Title:   serverTitle,
		},
		//nolint:exhaustruct // optional fields use defaults
		&mcp.ServerOptions{
			Instructions: systemPrompt,
		},
	)

	for _, r := range registrators {
		if err := r.Register(mcpServer); err != nil {
			return nil, fmt.Errorf("register tools: %w", err)
		}
	}

	return &Server{mcpServer: mcpServer}, nil
}

// Run starts the server on the given transport.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.mcpServer.Run(ctx, transport)
}

// Connect attaches the server to a transport for testing.
func (s *Server) Connect(ctx context.Context, transport mcp.Transport) (*mcp.ServerSession, error) {
	return s.mcpServer.Connect(ctx, transport, nil)
}

// MakeHandler adapts a tool function to the mcp.AddTool signature.
func MakeHandler[In, Out any](
	fn func(context.Context, In) (*Out, error),
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, *Out, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input In) (*mcp.CallToolResult, *Out, error) {
		output, err := fn(ctx, input)
		return nil, output, err
	}
}
