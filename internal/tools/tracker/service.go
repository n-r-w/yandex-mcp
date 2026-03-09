package tracker

import (
	"context"

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

// registerTool adds a tracker tool only when it is enabled in the registrator.
func registerTool[In, Out any](
	enabled bool,
	srv *mcp.Server,
	definition *mcp.Tool,
	handler func(context.Context, In) (*Out, error),
) {
	if !enabled {
		return
	}

	mcp.AddTool(srv, definition, server.MakeHandler(handler))
}

// Register registers all tracker tools with the MCP server.
func (r *Registrator) Register(srv *mcp.Server) error {
	registerTool(
		r.enabledTools[domain.TrackerToolIssueGet],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueGet.String(),
			Description: "Retrieves a Yandex Tracker issue by its ID or key",
		},
		r.getIssue,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolIssueSearch],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueSearch.String(),
			Description: "Searches Yandex Tracker issues using filter or query",
		},
		r.searchIssues,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolIssueCount],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolIssueCount.String(),
			Description: "Counts Yandex Tracker issues matching filter or query",
		},
		r.countIssues,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolTransitionsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolTransitionsList.String(),
			Description: "Lists available status transitions for a Yandex Tracker issue",
		},
		r.listTransitions,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolQueuesList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueuesList.String(),
			Description: "Lists Yandex Tracker queues",
		},
		r.listQueues,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolBoardsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolBoardsList.String(),
			Description: "Lists Yandex Tracker boards",
			InputSchema: emptyObjectInputSchema(),
		},
		r.listBoards,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolBoardSprintsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolBoardSprintsList.String(),
			Description: "Lists sprints for a Yandex Tracker board",
		},
		r.listBoardSprints,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolCommentsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker issue",
		},
		r.listComments,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolAttachmentsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentsList.String(),
			Description: "Lists attachments for a Yandex Tracker issue",
		},
		r.listAttachments,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolAttachmentGet],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentGet.String(),
			Description: "Downloads a file attached to a Yandex Tracker issue; requires exactly one of save_path or get_content",
		},
		r.getAttachment,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolAttachmentPreviewGet],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolAttachmentPreviewGet.String(),
			Description: "Downloads a thumbnail for a Yandex Tracker issue attachment",
		},
		r.getAttachmentPreview,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolQueueGet],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolQueueGet.String(),
			Description: "Gets a Yandex Tracker queue by ID or key",
		},
		r.getQueue,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolUserCurrent],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUserCurrent.String(),
			Description: "Gets the current authenticated Yandex Tracker user",
			InputSchema: emptyObjectInputSchema(),
		},
		r.getCurrentUser,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolUsersList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUsersList.String(),
			Description: "Lists Yandex Tracker users",
		},
		r.listUsers,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolUserGet],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolUserGet.String(),
			Description: "Gets a Yandex Tracker user by ID or login",
		},
		r.getUser,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolLinksList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolLinksList.String(),
			Description: "Lists all links for a Yandex Tracker issue",
		},
		r.listLinks,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolChangelog],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolChangelog.String(),
			Description: "Gets the changelog for a Yandex Tracker issue",
		},
		r.getChangelog,
	)
	registerTool(
		r.enabledTools[domain.TrackerToolProjectCommentsList],
		srv,
		&mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
			Name:        domain.TrackerToolProjectCommentsList.String(),
			Description: "Lists comments for a Yandex Tracker project entity",
		},
		r.listProjectComments,
	)

	return nil
}
