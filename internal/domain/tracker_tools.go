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
	TrackerToolCount // used to verify list completeness
)

// String returns the MCP tool name for the TrackerTool.
func (t TrackerTool) String() string {
	switch t {
	case TrackerToolIssueGet:
		return "tracker_issue_get"
	case TrackerToolIssueSearch:
		return "tracker_issue_search"
	case TrackerToolIssueCount:
		return "tracker_issue_count"
	case TrackerToolTransitionsList:
		return "tracker_issue_transitions_list"
	case TrackerToolQueuesList:
		return "tracker_queues_list"
	case TrackerToolCommentsList:
		return "tracker_issue_comments_list"
	case TrackerToolIssueCreate:
		return "tracker_issue_create"
	case TrackerToolIssueUpdate:
		return "tracker_issue_update"
	case TrackerToolTransitionExecute:
		return "tracker_issue_transition_execute"
	case TrackerToolCommentAdd:
		return "tracker_issue_comment_add"
	case TrackerToolCount:
		return ""
	default:
		return ""
	}
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
	}
}

// TrackerWriteTools returns the write tools enabled via --tracker-write flag.
func TrackerWriteTools() []TrackerTool {
	return []TrackerTool{
		TrackerToolIssueCreate,
		TrackerToolIssueUpdate,
		TrackerToolTransitionExecute,
		TrackerToolCommentAdd,
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
		TrackerToolIssueCreate,
		TrackerToolIssueUpdate,
		TrackerToolTransitionExecute,
		TrackerToolCommentAdd,
	}
}
