package wiki

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
)

// Registrator registers wiki tools with an MCP server.
type Registrator struct {
	adapter      IWikiAdapter
	enabledTools map[domain.WikiTool]bool
}

// Compile-time assertion that Registrator implements server.IToolsRegistrator.
var _ server.IToolsRegistrator = (*Registrator)(nil)

// NewRegistrator creates a new wiki tools registrator.
func NewRegistrator(adapter IWikiAdapter, enabledTools []domain.WikiTool) *Registrator {
	toolMap := make(map[domain.WikiTool]bool, len(enabledTools))
	for _, t := range enabledTools {
		toolMap[t] = true
	}

	return &Registrator{
		adapter:      adapter,
		enabledTools: toolMap,
	}
}

// Register registers all wiki tools with the MCP server.
func (r *Registrator) Register(srv *mcp.Server) error {
	if r.enabledTools[domain.WikiToolPageGetBySlug] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageGetBySlug.String(),
			Description: "Retrieves a Yandex Wiki page by its slug (URL path)",
		}, server.MakeHandler(r.getPageBySlug))
	}

	if r.enabledTools[domain.WikiToolPageGetByID] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageGetByID.String(),
			Description: "Retrieves a Yandex Wiki page by its numeric ID",
		}, server.MakeHandler(r.getPageByID))
	}

	if r.enabledTools[domain.WikiToolResourcesList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolResourcesList.String(),
			Description: "Lists resources (attachments, grids) for a Yandex Wiki page",
		}, server.MakeHandler(r.listResources))
	}

	if r.enabledTools[domain.WikiToolGridsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridsList.String(),
			Description: "Lists dynamic tables (grids) for a Yandex Wiki page",
		}, server.MakeHandler(r.listGrids))
	}

	if r.enabledTools[domain.WikiToolGridGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridGet.String(),
			Description: "Retrieves a Yandex Wiki dynamic table (grid) by its ID",
		}, server.MakeHandler(r.getGrid))
	}

	return nil
}
