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

	if r.enabledTools[domain.WikiToolPageCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageCreate.String(),
			Description: "Creates a new Yandex Wiki page",
		}, server.MakeHandler(r.createPage))
	}

	if r.enabledTools[domain.WikiToolPageUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageUpdate.String(),
			Description: "Updates an existing Yandex Wiki page",
		}, server.MakeHandler(r.updatePage))
	}

	if r.enabledTools[domain.WikiToolPageAppend] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageAppend.String(),
			Description: "Appends content to an existing Yandex Wiki page",
		}, server.MakeHandler(r.appendPage))
	}

	if r.enabledTools[domain.WikiToolGridCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridCreate.String(),
			Description: "Creates a new Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.createGrid))
	}

	if r.enabledTools[domain.WikiToolGridCellsUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridCellsUpdate.String(),
			Description: "Updates cells in a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.updateGridCells))
	}

	if r.enabledTools[domain.WikiToolPageDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageDelete.String(),
			Description: "Deletes a Yandex Wiki page and returns a recovery token",
		}, server.MakeHandler(r.deletePage))
	}

	if r.enabledTools[domain.WikiToolPageClone] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolPageClone.String(),
			Description: "Clones a Yandex Wiki page to a new location (async operation)",
		}, server.MakeHandler(r.clonePage))
	}

	if r.enabledTools[domain.WikiToolGridDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridDelete.String(),
			Description: "Deletes a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.deleteGrid))
	}

	if r.enabledTools[domain.WikiToolGridClone] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridClone.String(),
			Description: "Clones a Yandex Wiki grid to a new location (async operation)",
		}, server.MakeHandler(r.cloneGrid))
	}

	if r.enabledTools[domain.WikiToolGridRowsAdd] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridRowsAdd.String(),
			Description: "Adds rows to a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.addGridRows))
	}

	if r.enabledTools[domain.WikiToolGridRowsDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridRowsDelete.String(),
			Description: "Deletes rows from a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.deleteGridRows))
	}

	if r.enabledTools[domain.WikiToolGridRowsMove] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridRowsMove.String(),
			Description: "Moves rows within a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.moveGridRows))
	}

	if r.enabledTools[domain.WikiToolGridColumnsAdd] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridColumnsAdd.String(),
			Description: "Adds columns to a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.addGridColumns))
	}

	if r.enabledTools[domain.WikiToolGridColumnsDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridColumnsDelete.String(),
			Description: "Deletes columns from a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.deleteGridColumns))
	}

	if r.enabledTools[domain.WikiToolGridColumnsMove] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.WikiToolGridColumnsMove.String(),
			Description: "Moves columns within a Yandex Wiki dynamic table (grid)",
		}, server.MakeHandler(r.moveGridColumns))
	}

	return nil
}

func (r *Registrator) logError(ctx context.Context, err error) error {
	return domain.LogError(ctx, string(domain.ServiceWiki), err)
}
