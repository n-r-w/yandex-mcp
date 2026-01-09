package wiki

import (
	"context"

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
		}, server.MakeHandler(r.GetPageBySlug))
	}

	if r.enabledTools[domain.WikiToolPageGetByID] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageGetByID.String(),
			Description: "Retrieves a Yandex Wiki page by its numeric ID",
		}, server.MakeHandler(r.GetPageByID))
	}

	if r.enabledTools[domain.WikiToolResourcesList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolResourcesList.String(),
			Description: "Lists resources (attachments, grids) for a Yandex Wiki page",
		}, server.MakeHandler(r.ListResources))
	}

	if r.enabledTools[domain.WikiToolGridsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridsList.String(),
			Description: "Lists dynamic tables (grids) for a Yandex Wiki page",
		}, server.MakeHandler(r.ListGrids))
	}

	if r.enabledTools[domain.WikiToolGridGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridGet.String(),
			Description: "Retrieves a Yandex Wiki dynamic table (grid) by its ID",
		}, server.MakeHandler(r.GetGrid))
	}

	if r.enabledTools[domain.WikiToolPageCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageCreate.String(),
			Description: "Creates a new Yandex Wiki page",
		}, server.MakeHandler(r.CreatePage))
	}

	if r.enabledTools[domain.WikiToolPageUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageUpdate.String(),
			Description: "Updates an existing Yandex Wiki page",
		}, server.MakeHandler(r.UpdatePage))
	}

	if r.enabledTools[domain.WikiToolPageAppend] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageAppend.String(),
			Description: "Appends content to an existing Yandex Wiki page",
		}, server.MakeHandler(r.AppendPage))
	}

	if r.enabledTools[domain.WikiToolGridCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridCreate.String(),
			Description: "Creates a new Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.CreateGrid))
	}

	if r.enabledTools[domain.WikiToolGridCellsUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridCellsUpdate.String(),
			Description: "Updates cells in a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.UpdateGridCells))
	}

	return nil
}

func (r *Registrator) logError(ctx context.Context, err error) error {
	return domain.LogError(ctx, string(domain.ServiceWiki), err)
}
