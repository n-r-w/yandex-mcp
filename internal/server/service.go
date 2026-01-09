package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and handles tool registration.
type Server struct {
	mcpServer *mcp.Server
}

// New creates a new MCP server with the provided tool registrators.
// Each registrator is responsible for registering its own tools with the MCP server.
func New(registrators []IToolsRegistrator) (*Server, error) {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{ //nolint:exhaustruct // optional fields use defaults
			Name:    serverName,
			Version: serverVersion,
			Title:   setverTitle,
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

// Run starts the MCP server using the provided transport.
// This method blocks until the context is cancelled or an error occurs.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.mcpServer.Run(ctx, transport)
}

// Connect connects the server to a transport.
// Useful for testing with in-memory transports.
func (s *Server) Connect(ctx context.Context, transport mcp.Transport) (*mcp.ServerSession, error) {
	return s.mcpServer.Connect(ctx, transport, nil)
}
