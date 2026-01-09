package tracker

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
)

// Registrator registers tracker tools with an MCP server.
type Registrator struct {
	adapter      ITrackerAdapter
	enabledTools map[domain.TrackerTool]bool
}

// Compile-time assertion that Registrator implements server.IToolsRegistrator.
var _ server.IToolsRegistrator = (*Registrator)(nil)

// NewRegistrator creates a new tracker tools registrator.
func NewRegistrator(adapter ITrackerAdapter, enabledTools []domain.TrackerTool) *Registrator {
	toolMap := make(map[domain.TrackerTool]bool, len(enabledTools))
	for _, t := range enabledTools {
		toolMap[t] = true
	}

	return &Registrator{
		adapter:      adapter,
		enabledTools: toolMap,
	}
}

// Register registers all tracker tools with the MCP server.
func (r *Registrator) Register(srv *mcp.Server) error {
	if r.enabledTools[domain.TrackerToolIssueGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueGet.String(),
			Description: "Retrieves a Yandex Tracker issue by its ID or key",
		}, server.MakeHandler(r.GetIssue))
	}

	if r.enabledTools[domain.TrackerToolIssueSearch] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueSearch.String(),
			Description: "Searches Yandex Tracker issues using filter or query",
		}, server.MakeHandler(r.SearchIssues))
	}

	if r.enabledTools[domain.TrackerToolIssueCount] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueCount.String(),
			Description: "Counts Yandex Tracker issues matching filter or query",
		}, server.MakeHandler(r.CountIssues))
	}

	if r.enabledTools[domain.TrackerToolTransitionsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolTransitionsList.String(),
			Description: "Lists available status transitions for a Yandex Tracker issue",
		}, server.MakeHandler(r.ListTransitions))
	}

	if r.enabledTools[domain.TrackerToolQueuesList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueuesList.String(),
			Description: "Lists Yandex Tracker queues",
		}, server.MakeHandler(r.ListQueues))
	}

	if r.enabledTools[domain.TrackerToolCommentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker issue",
		}, server.MakeHandler(r.ListComments))
	}

	if r.enabledTools[domain.TrackerToolIssueCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueCreate.String(),
			Description: "Creates a new Yandex Tracker issue",
		}, server.MakeHandler(r.CreateIssue))
	}

	if r.enabledTools[domain.TrackerToolIssueUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueUpdate.String(),
			Description: "Updates an existing Yandex Tracker issue",
		}, server.MakeHandler(r.UpdateIssue))
	}

	if r.enabledTools[domain.TrackerToolTransitionExecute] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolTransitionExecute.String(),
			Description: "Executes a status transition on a Yandex Tracker issue",
		}, server.MakeHandler(r.ExecuteTransition))
	}

	if r.enabledTools[domain.TrackerToolCommentAdd] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentAdd.String(),
			Description: "Adds a comment to a Yandex Tracker issue",
		}, server.MakeHandler(r.AddComment))
	}

	return nil
}
