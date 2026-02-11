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
	}
}
