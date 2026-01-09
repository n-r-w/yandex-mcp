// Package server provides MCP server bootstrap and tool registration.
package server

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=server

import "github.com/modelcontextprotocol/go-sdk/mcp"

// IToolsRegistrator defines the contract for registering tools with an MCP server.
// Implementations are provided by tool packages and passed via dependency injection.
type IToolsRegistrator interface {
	Register(srv *mcp.Server) error
}
