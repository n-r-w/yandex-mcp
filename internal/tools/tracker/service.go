package tracker

import (
	"context"

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
		}, server.MakeHandler(r.getIssue))
	}

	if r.enabledTools[domain.TrackerToolIssueSearch] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueSearch.String(),
			Description: "Searches Yandex Tracker issues using filter or query",
		}, server.MakeHandler(r.searchIssues))
	}

	if r.enabledTools[domain.TrackerToolIssueCount] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueCount.String(),
			Description: "Counts Yandex Tracker issues matching filter or query",
		}, server.MakeHandler(r.countIssues))
	}

	if r.enabledTools[domain.TrackerToolTransitionsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolTransitionsList.String(),
			Description: "Lists available status transitions for a Yandex Tracker issue",
		}, server.MakeHandler(r.listTransitions))
	}

	if r.enabledTools[domain.TrackerToolQueuesList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueuesList.String(),
			Description: "Lists Yandex Tracker queues",
		}, server.MakeHandler(r.listQueues))
	}

	if r.enabledTools[domain.TrackerToolCommentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker issue",
		}, server.MakeHandler(r.listComments))
	}

	if r.enabledTools[domain.TrackerToolIssueCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueCreate.String(),
			Description: "Creates a new Yandex Tracker issue",
		}, server.MakeHandler(r.createIssue))
	}

	if r.enabledTools[domain.TrackerToolIssueUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueUpdate.String(),
			Description: "Updates an existing Yandex Tracker issue",
		}, server.MakeHandler(r.updateIssue))
	}

	if r.enabledTools[domain.TrackerToolTransitionExecute] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolTransitionExecute.String(),
			Description: "Executes a status transition on a Yandex Tracker issue",
		}, server.MakeHandler(r.executeTransition))
	}

	if r.enabledTools[domain.TrackerToolCommentAdd] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentAdd.String(),
			Description: "Adds a comment to a Yandex Tracker issue",
		}, server.MakeHandler(r.addComment))
	}

	if r.enabledTools[domain.TrackerToolCommentUpdate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentUpdate.String(),
			Description: "Updates an existing comment on a Yandex Tracker issue",
		}, server.MakeHandler(r.updateComment))
	}

	if r.enabledTools[domain.TrackerToolCommentDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentDelete.String(),
			Description: "Deletes a comment from a Yandex Tracker issue",
		}, server.MakeHandler(r.deleteComment))
	}

	if r.enabledTools[domain.TrackerToolAttachmentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentsList.String(),
			Description: "Lists attachments for a Yandex Tracker issue",
		}, server.MakeHandler(r.listAttachments))
	}

	if r.enabledTools[domain.TrackerToolAttachmentDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentDelete.String(),
			Description: "Deletes an attachment from a Yandex Tracker issue",
		}, server.MakeHandler(r.deleteAttachment))
	}

	if r.enabledTools[domain.TrackerToolQueueGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueGet.String(),
			Description: "Gets a Yandex Tracker queue by ID or key",
		}, server.MakeHandler(r.getQueue))
	}

	if r.enabledTools[domain.TrackerToolQueueCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueCreate.String(),
			Description: "Creates a new Yandex Tracker queue",
		}, server.MakeHandler(r.createQueue))
	}

	if r.enabledTools[domain.TrackerToolQueueDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueDelete.String(),
			Description: "Deletes a Yandex Tracker queue",
		}, server.MakeHandler(r.deleteQueue))
	}

	if r.enabledTools[domain.TrackerToolQueueRestore] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueRestore.String(),
			Description: "Restores a deleted Yandex Tracker queue",
		}, server.MakeHandler(r.restoreQueue))
	}

	if r.enabledTools[domain.TrackerToolUserCurrent] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUserCurrent.String(),
			Description: "Gets the current authenticated Yandex Tracker user",
		}, server.MakeHandler(r.getCurrentUser))
	}

	if r.enabledTools[domain.TrackerToolUsersList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUsersList.String(),
			Description: "Lists Yandex Tracker users",
		}, server.MakeHandler(r.listUsers))
	}

	if r.enabledTools[domain.TrackerToolUserGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUserGet.String(),
			Description: "Gets a Yandex Tracker user by ID or login",
		}, server.MakeHandler(r.getUser))
	}

	if r.enabledTools[domain.TrackerToolLinksList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolLinksList.String(),
			Description: "Lists all links for a Yandex Tracker issue",
		}, server.MakeHandler(r.listLinks))
	}

	if r.enabledTools[domain.TrackerToolLinkCreate] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolLinkCreate.String(),
			Description: "Creates a link between Yandex Tracker issues",
		}, server.MakeHandler(r.createLink))
	}

	if r.enabledTools[domain.TrackerToolLinkDelete] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolLinkDelete.String(),
			Description: "Deletes a link from a Yandex Tracker issue",
		}, server.MakeHandler(r.deleteLink))
	}

	if r.enabledTools[domain.TrackerToolChangelog] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolChangelog.String(),
			Description: "Gets the changelog for a Yandex Tracker issue",
		}, server.MakeHandler(r.getChangelog))
	}

	if r.enabledTools[domain.TrackerToolIssueMove] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueMove.String(),
			Description: "Moves a Yandex Tracker issue to another queue",
		}, server.MakeHandler(r.moveIssue))
	}

	if r.enabledTools[domain.TrackerToolProjectCommentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolProjectCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker project entity",
		}, server.MakeHandler(r.listProjectComments))
	}

	return nil
}

func (r *Registrator) logError(ctx context.Context, err error) error {
	return domain.LogError(ctx, string(domain.ServiceTracker), err)
}
