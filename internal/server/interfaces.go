// Package server provides MCP server bootstrap and tool registration.
package server

//go:generate go tool mockgen -source=interfaces.go -destination=mock_interfaces.go -package=server

import "github.com/modelcontextprotocol/go-sdk/mcp"

// IToolsRegistrator abstracts tool registration for dependency injection.
type IToolsRegistrator interface {
	Register(srv *mcp.Server) error
}
