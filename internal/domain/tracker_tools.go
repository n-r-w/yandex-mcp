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
	TrackerToolCommentsList
	TrackerToolIssueCreate
	TrackerToolIssueUpdate
	TrackerToolTransitionExecute
	TrackerToolCommentAdd
	TrackerToolCommentUpdate
	TrackerToolCommentDelete
	TrackerToolAttachmentsList
	TrackerToolAttachmentDelete
	TrackerToolQueueGet
	TrackerToolQueueCreate
	TrackerToolQueueDelete
	TrackerToolQueueRestore
	TrackerToolUserCurrent
	TrackerToolUsersList
	TrackerToolUserGet
	TrackerToolLinksList
	TrackerToolLinkCreate
	TrackerToolLinkDelete
	TrackerToolChangelog
	TrackerToolIssueMove
	TrackerToolProjectCommentsList
	TrackerToolCount // used to verify list completeness
)

// String returns the MCP tool name for the TrackerTool.
func (t TrackerTool) String() string {
	names := map[TrackerTool]string{
		TrackerToolIssueGet:            "tracker_issue_get",
		TrackerToolIssueSearch:         "tracker_issue_search",
		TrackerToolIssueCount:          "tracker_issue_count",
		TrackerToolTransitionsList:     "tracker_issue_transitions_list",
		TrackerToolQueuesList:          "tracker_queues_list",
		TrackerToolCommentsList:        "tracker_issue_comments_list",
		TrackerToolIssueCreate:         "tracker_issue_create",
		TrackerToolIssueUpdate:         "tracker_issue_update",
		TrackerToolTransitionExecute:   "tracker_issue_transition_execute",
		TrackerToolCommentAdd:          "tracker_issue_comment_add",
		TrackerToolCommentUpdate:       "tracker_issue_comment_update",
		TrackerToolCommentDelete:       "tracker_issue_comment_delete",
		TrackerToolAttachmentsList:     "tracker_issue_attachments_list",
		TrackerToolAttachmentDelete:    "tracker_issue_attachment_delete",
		TrackerToolQueueGet:            "tracker_queue_get",
		TrackerToolQueueCreate:         "tracker_queue_create",
		TrackerToolQueueDelete:         "tracker_queue_delete",
		TrackerToolQueueRestore:        "tracker_queue_restore",
		TrackerToolUserCurrent:         "tracker_user_current",
		TrackerToolUsersList:           "tracker_users_list",
		TrackerToolUserGet:             "tracker_user_get",
		TrackerToolLinksList:           "tracker_issue_links_list",
		TrackerToolLinkCreate:          "tracker_issue_link_create",
		TrackerToolLinkDelete:          "tracker_issue_link_delete",
		TrackerToolChangelog:           "tracker_issue_changelog",
		TrackerToolIssueMove:           "tracker_issue_move",
		TrackerToolProjectCommentsList: "tracker_project_comments_list",
	}
	return names[t]
}

// TrackerReadOnlyTools returns the default read-only tools.
func TrackerReadOnlyTools() []TrackerTool {
	return []TrackerTool{
		TrackerToolIssueGet,
		TrackerToolIssueSearch,
		TrackerToolIssueCount,
		TrackerToolTransitionsList,
		TrackerToolQueuesList,
		TrackerToolCommentsList,
		TrackerToolAttachmentsList,
		TrackerToolQueueGet,
		TrackerToolUserCurrent,
		TrackerToolUsersList,
		TrackerToolUserGet,
		TrackerToolLinksList,
		TrackerToolChangelog,
		TrackerToolProjectCommentsList,
	}
}

// TrackerWriteTools returns the write tools enabled via --tracker-write flag.
func TrackerWriteTools() []TrackerTool {
	return []TrackerTool{
		TrackerToolIssueCreate,
		TrackerToolIssueUpdate,
		TrackerToolTransitionExecute,
		TrackerToolCommentAdd,
		TrackerToolCommentUpdate,
		TrackerToolCommentDelete,
		TrackerToolAttachmentDelete,
		TrackerToolQueueCreate,
		TrackerToolQueueDelete,
		TrackerToolQueueRestore,
		TrackerToolLinkCreate,
		TrackerToolLinkDelete,
		TrackerToolIssueMove,
	}
}

// TrackerAllTools returns all tracker tools in stable order.
func TrackerAllTools() []TrackerTool {
	return []TrackerTool{
		TrackerToolIssueGet,
		TrackerToolIssueSearch,
		TrackerToolIssueCount,
		TrackerToolTransitionsList,
		TrackerToolQueuesList,
		TrackerToolCommentsList,
		TrackerToolAttachmentsList,
		TrackerToolQueueGet,
		TrackerToolUserCurrent,
		TrackerToolUsersList,
		TrackerToolUserGet,
		TrackerToolIssueCreate,
		TrackerToolIssueUpdate,
		TrackerToolTransitionExecute,
		TrackerToolCommentAdd,
		TrackerToolCommentUpdate,
		TrackerToolCommentDelete,
		TrackerToolAttachmentDelete,
		TrackerToolQueueCreate,
		TrackerToolQueueDelete,
		TrackerToolQueueRestore,
		TrackerToolLinksList,
		TrackerToolLinkCreate,
		TrackerToolLinkDelete,
		TrackerToolChangelog,
		TrackerToolIssueMove,
		TrackerToolProjectCommentsList,
	}
}
