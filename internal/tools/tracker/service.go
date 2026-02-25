package tracker

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
)

// Registrator registers tracker tools with an MCP server.
type Registrator struct {
	adapter           ITrackerAdapter
	enabledTools      map[domain.TrackerTool]bool
	allowedExtensions []string
	allowedViewExts   []string
	allowedDirs       []string
}

// Compile-time assertion that Registrator implements server.IToolsRegistrator.
var _ server.IToolsRegistrator = (*Registrator)(nil)

// NewRegistrator creates a new tracker tools registrator.
func NewRegistrator(
	adapter ITrackerAdapter,
	enabledTools []domain.TrackerTool,
	allowedExtensions []string,
	allowedViewExts []string,
	allowedDirs []string,
) *Registrator {
	toolMap := make(map[domain.TrackerTool]bool, len(enabledTools))
	for _, t := range enabledTools {
		toolMap[t] = true
	}

	return &Registrator{
		adapter:           adapter,
		enabledTools:      toolMap,
		allowedExtensions: normalizeAllowedExtensions(allowedExtensions),
		allowedViewExts:   normalizeAllowedExtensions(allowedViewExts),
		allowedDirs:       normalizeAllowedDirs(allowedDirs),
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

	if r.enabledTools[domain.TrackerToolAttachmentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentsList.String(),
			Description: "Lists attachments for a Yandex Tracker issue",
		}, server.MakeHandler(r.listAttachments))
	}

	if r.enabledTools[domain.TrackerToolAttachmentGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentGet.String(),
			Description: "Downloads a file attached to a Yandex Tracker issue; requires exactly one of save_path or get_content",
		}, server.MakeHandler(r.getAttachment))
	}

	if r.enabledTools[domain.TrackerToolAttachmentPreviewGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentPreviewGet.String(),
			Description: "Downloads a thumbnail for a Yandex Tracker issue attachment",
		}, server.MakeHandler(r.getAttachmentPreview))
	}

	if r.enabledTools[domain.TrackerToolQueueGet] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueGet.String(),
			Description: "Gets a Yandex Tracker queue by ID or key",
		}, server.MakeHandler(r.getQueue))
	}

	if r.enabledTools[domain.TrackerToolUserCurrent] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUserCurrent.String(),
			Description: "Gets the current authenticated Yandex Tracker user",
			InputSchema: emptyObjectInputSchema(),
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

	if r.enabledTools[domain.TrackerToolChangelog] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolChangelog.String(),
			Description: "Gets the changelog for a Yandex Tracker issue",
		}, server.MakeHandler(r.getChangelog))
	}

	if r.enabledTools[domain.TrackerToolProjectCommentsList] {
		mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolProjectCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker project entity",
		}, server.MakeHandler(r.listProjectComments))
	}

	return nil
}
