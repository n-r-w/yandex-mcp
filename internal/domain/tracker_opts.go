package domain

// TrackerGetIssueOpts represents options for getting a single issue.
type TrackerGetIssueOpts struct {
	Expand string
}

// TrackerSearchIssuesOpts represents options for searching issues.
type TrackerSearchIssuesOpts struct {
	// Filter is a field-based filter as key-value string pairs.
	// The adapter layer converts this to the appropriate JSON structure.
	Filter map[string]string
	// Query is a query language filter string.
	Query string
	// Order specifies sorting direction and field.
	Order string
	// Expand specifies additional fields to include.
	Expand string
	// PerPage specifies the number of results per page.
	PerPage int
	// Page specifies the page number.
	Page int
	// ScrollType specifies scrolling type.
	ScrollType string
	// PerScroll specifies max issues per response in scroll mode.
	PerScroll int
	// ScrollTTLMillis specifies scroll context lifetime in milliseconds.
	ScrollTTLMillis int
	// ScrollID specifies the scroll page ID for subsequent requests.
	ScrollID string
}

// TrackerCountIssuesOpts represents options for counting issues.
type TrackerCountIssuesOpts struct {
	// Filter is a field-based filter as key-value string pairs.
	Filter map[string]string
	// Query is a query language filter string.
	Query string
}

// TrackerListQueuesOpts represents options for listing queues.
type TrackerListQueuesOpts struct {
	// Expand specifies additional fields to include.
	Expand string
	// PerPage specifies the number of queues per page.
	PerPage int
	// Page specifies the page number.
	Page int
}

// TrackerListCommentsOpts represents options for listing comments.
type TrackerListCommentsOpts struct {
	// Expand specifies additional fields to include.
	Expand string
	// PerPage specifies the number of comments per page.
	PerPage int
	// ID specifies the comment ID after which the requested page begins.
	ID string
}

// TrackerGetQueueOpts represents options for getting a single queue.
type TrackerGetQueueOpts struct {
	// Expand specifies additional fields to include.
	// Allowed values: projects, components, versions, types, team, workflows, all.
	Expand string
}

// TrackerListUsersOpts represents options for listing users.
type TrackerListUsersOpts struct {
	// PerPage specifies the number of users per page.
	PerPage int
	// Page specifies the page number.
	Page int
}

// TrackerGetChangelogOpts represents options for getting issue changelog.
type TrackerGetChangelogOpts struct {
	// PerPage specifies the number of changelog entries per page.
	// Default: 50.
	PerPage int
}

// TrackerListProjectCommentsOpts represents options for listing project comments.
type TrackerListProjectCommentsOpts struct {
	// Expand specifies additional fields to include.
	// Allowed values: all, html, attachments, reactions.
	Expand string
}
