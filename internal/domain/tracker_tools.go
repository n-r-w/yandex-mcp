package domain

// TrackerTool represents a specific Tracker tool that can be registered.
type TrackerTool int

// TrackerTool constants represent individual Tracker tools.
const (
	TrackerToolIssueGet TrackerTool = iota
	TrackerToolIssueSearch
	TrackerToolIssueCount
	TrackerToolTransitionsList
	TrackerToolQueuesList
	TrackerToolBoardsList
	TrackerToolBoardSprintsList
	TrackerToolCommentsList
	TrackerToolAttachmentsList
	TrackerToolAttachmentGet
	TrackerToolAttachmentPreviewGet
	TrackerToolQueueGet
	TrackerToolUserCurrent
	TrackerToolUsersList
	TrackerToolUserGet
	TrackerToolLinksList
	TrackerToolChangelog
	TrackerToolProjectCommentsList
	TrackerToolIssueCreate
	TrackerToolCount // used to verify list completeness
)

// String returns the MCP tool name for the TrackerTool.
func (t TrackerTool) String() string {
	names := map[TrackerTool]string{
		TrackerToolIssueGet:             "tracker_issue_get",
		TrackerToolIssueSearch:          "tracker_issue_search",
		TrackerToolIssueCount:           "tracker_issue_count",
		TrackerToolTransitionsList:      "tracker_issue_transitions_list",
		TrackerToolQueuesList:           "tracker_queues_list",
		TrackerToolBoardsList:           "tracker_boards_list",
		TrackerToolBoardSprintsList:     "tracker_board_sprints_list",
		TrackerToolCommentsList:         "tracker_issue_comments_list",
		TrackerToolAttachmentsList:      "tracker_issue_attachments_list",
		TrackerToolAttachmentGet:        "tracker_issue_attachment_get",
		TrackerToolAttachmentPreviewGet: "tracker_issue_attachment_preview_get",
		TrackerToolQueueGet:             "tracker_queue_get",
		TrackerToolUserCurrent:          "tracker_user_current",
		TrackerToolUsersList:            "tracker_users_list",
		TrackerToolUserGet:              "tracker_user_get",
		TrackerToolLinksList:            "tracker_issue_links_list",
		TrackerToolChangelog:            "tracker_issue_changelog",
		TrackerToolProjectCommentsList:  "tracker_project_comments_list",
		TrackerToolIssueCreate:          "tracker_issue_create",
	}
	return names[t]
}

// TrackerAllTools returns all tracker tools in stable order.
func TrackerAllTools() []TrackerTool {
	return []TrackerTool{
		TrackerToolIssueGet,
		TrackerToolIssueSearch,
		TrackerToolIssueCount,
		TrackerToolTransitionsList,
		TrackerToolQueuesList,
		TrackerToolBoardsList,
		TrackerToolBoardSprintsList,
		TrackerToolCommentsList,
		TrackerToolAttachmentsList,
		TrackerToolAttachmentGet,
		TrackerToolAttachmentPreviewGet,
		TrackerToolQueueGet,
		TrackerToolUserCurrent,
		TrackerToolUsersList,
		TrackerToolUserGet,
		TrackerToolLinksList,
		TrackerToolChangelog,
		TrackerToolProjectCommentsList,
		TrackerToolIssueCreate,
	}
}

// ToolGroup classifies tools as read-only or write (mutating).
type ToolGroup int

const (
	// ToolGroupRead marks tools that only read data.
	ToolGroupRead ToolGroup = iota
	// ToolGroupWrite marks tools that create, update, or delete data.
	ToolGroupWrite
)

// Group returns the tool group for the tracker tool.
func (t TrackerTool) Group() ToolGroup {
	switch t {
	case TrackerToolIssueCreate:
		return ToolGroupWrite
	default:
		return ToolGroupRead
	}
}

// TrackerReadTools returns only read-only tracker tools.
func TrackerReadTools() []TrackerTool {
	var tools []TrackerTool
	for _, t := range TrackerAllTools() {
		if t.Group() == ToolGroupRead {
			tools = append(tools, t)
		}
	}
	return tools
}

// TrackerWriteTools returns only write (mutating) tracker tools.
func TrackerWriteTools() []TrackerTool {
	var tools []TrackerTool
	for _, t := range TrackerAllTools() {
		if t.Group() == ToolGroupWrite {
			tools = append(tools, t)
		}
	}
	return tools
}
