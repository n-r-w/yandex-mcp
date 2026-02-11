//nolint:lll // JSON schema descriptions for LLM tool inputs require detailed inline documentation
package tracker

// Input DTOs for tracker tools.

// getIssueInputDTO is the input for tracker_issue_get tool.
type getIssueInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'attachments' (attached files metadata). Example: 'attachments'"`
}

// searchIssuesInputDTO is the input for tracker_issue_search tool.
type searchIssuesInputDTO struct {
	// Filter is a field-based filter object with key-value pairs.
	// All values are strings. Multiple values are comma-separated.
	// Examples:
	//   - Single value: {"queue": "TREK"}
	//   - Multiple values: {"status": "Open,In Progress"}
	//   - Special functions: {"assignee": "me()"}, {"assignee": "empty()"}
	//   - Combined: {"queue": "CP", "assignee": "me()", "status": "Open,In Progress"}
	// NOTE: Cannot be used together with 'query' parameter.
	Filter map[string]string `json:"filter,omitempty" jsonschema:"Field-based filter with key-value pairs. Values: simple values, special functions (me(), empty()), or comma-separated multiple values. Examples: {\"queue\": \"CP\"}, {\"status\": \"Open,In Progress\"}, {\"assignee\": \"me()\"}, {\"queue\": \"CP\", \"assignee\": \"me()\"}. IMPORTANT: Cannot be used together with 'query' - use either filter or query, not both."`

	// Query is a query language filter string using Yandex Tracker query syntax.
	// Supports complex boolean expressions with operators AND, OR, NOT, date functions.
	// NOTE: Cannot be used together with 'filter' parameter. Order parameter is ignored when using query.
	Query string `json:"query,omitempty" jsonschema:"Query language filter (Yandex Tracker syntax). Supports: field=value comparison, AND/OR/NOT operators, parentheses for grouping, date functions (today(), now(), today()-7d, today()+30d), special functions (me(), empty()). Supported fields: Queue, Status, Priority, Assignee, Author, Type, Resolution, Updated, Created, Due. Operators: : (exact match), >, <, >=, <= (numeric/dates). Examples: 'Status: Open', 'Assignee: me() AND Priority: Critical', '(Assignee: me() OR Author: me()) AND NOT Status: Closed', 'Updated: >today()-7d', 'Queue: CP OR BB AND NOT Status: Closed', 'Resolution: empty()'. IMPORTANT: Cannot be used together with 'filter' - use either filter or query, not both. Order parameter is ignored when using query."`

	// Order specifies sorting field and direction.
	// Format: [+/-]<field_key>
	// Examples: "+updated" (ascending), "-created" (descending)
	// NOTE: Only works with 'filter' parameter, ignored when using 'query'.
	Order string `json:"order,omitempty" jsonschema:"Issue sorting direction and field. Format: [+/-]<field_key>. Examples: '+updated' (ascending), '-created' (descending). Supported fields: created, updated, due, priority, status, id, key. IMPORTANT: Only works with 'filter' parameter, completely ignored when using 'query' parameter."`

	// Expand specifies additional fields to include in the response.
	// Possible values:
	//   - "transitions": workflow transitions between statuses
	//   - "attachments": attached files metadata
	// Can be combined: "transitions,attachments"
	Expand string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'transitions' (workflow transitions between statuses), 'attachments' (attached files metadata). Can be combined: 'transitions,attachments'. Example: 'transitions,attachments'"`

	// PerPage is the number of results per page for standard pagination.
	// Valid range: 1-50. Default: 50.
	// NOTE: For result sets < 10,000 issues. Use scroll for larger sets.
	PerPage int `json:"per_page,omitempty" jsonschema:"Number of results per page for standard pagination. Valid range: 1-50 (default: 50). Use for result sets < 10,000 issues. For larger sets (>10,000), use scroll pagination instead (per_scroll, scroll_id, scroll_type, scroll_ttl_millis)."`

	// Page is the page number for standard pagination (1-based).
	// Default: 1
	// NOTE: For result sets < 10,000 issues. Use scroll for larger sets.
	Page int `json:"page,omitempty" jsonschema:"Page number for standard pagination (1-based, default: 1). Use for result sets < 10,000 issues. For larger sets (>10,000), use scroll pagination instead (per_scroll, scroll_id, scroll_type, scroll_ttl_millis)."`

	// ScrollType determines scroll behavior for large result sets (>10,000 issues).
	// Possible values:
	//   - "sorted": use sorting specified in 'order' parameter
	//   - "unsorted": no sorting applied (faster)
	// NOTE: Only specified in first scroll request.
	ScrollType string `json:"scroll_type,omitempty" jsonschema:"Scroll type for large result sets (>10,000 issues). Possible values: 'sorted' (use sorting from 'order' parameter), 'unsorted' (no sorting, faster). Only specified in first scroll request. Use with: per_scroll, scroll_ttl_millis, scroll_id for subsequent requests."`

	// PerScroll is the maximum number of issues per scroll response.
	// Valid range: 1-1000. Default: 100.
	// NOTE: Only specified in first scroll request.
	PerScroll int `json:"per_scroll,omitempty" jsonschema:"Maximum number of issues per scroll response. Valid range: 1-1000 (default: 100). Only specified in first scroll request. Use for result sets >10,000 issues. Combine with: scroll_type, scroll_ttl_millis, and then use scroll_id for subsequent requests."`

	// ScrollTTLMillis is the scroll context lifetime in milliseconds.
	// Default: 60000 (60 seconds).
	// Maximum: 600000 (10 minutes).
	// NOTE: Only specified in first scroll request.
	ScrollTTLMillis int `json:"scroll_ttl_millis,omitempty" jsonschema:"Scroll context lifetime in milliseconds. Default: 60000 (60 seconds), maximum: 600000 (10 minutes). Only specified in first scroll request. After expiration, scroll_id becomes invalid and new scroll must be started. Use for result sets >10,000 issues."`

	// ScrollID is the scroll page identifier for pagination.
	// Returned in first scroll response, used in subsequent requests.
	// Example: "6962987e5d10fe1be1cacfa9"
	ScrollID string `json:"scroll_id,omitempty" jsonschema:"Scroll page identifier from previous scroll response. Use in 2nd and subsequent scroll requests to get next page of results. Obtained from 'scroll_id' field in first scroll response. Only for use with scroll pagination (>10,000 results). Example: '6962987e5d10fe1be1cacfa9'. Do not use with standard page/per_page pagination."`
}

// countIssuesInputDTO is the input for tracker_issue_count tool.
type countIssuesInputDTO struct {
	Filter map[string]string `json:"filter,omitempty" jsonschema:"Field-based filter with key-value pairs. Values: simple values, special functions (me(), empty()), or comma-separated multiple values. Examples: {\"queue\": \"CP\"}, {\"status\": \"Open,In Progress\"}, {\"assignee\": \"me()\"}. IMPORTANT: Cannot be used together with 'query' - use either filter or query, not both."`
	Query  string            `json:"query,omitempty" jsonschema:"Query language filter (Yandex Tracker syntax). Supports: field=value comparison, AND/OR/NOT operators, parentheses for grouping, date functions (today(), now(), today()-7d, today()+30d), special functions (me(), empty()). Supported fields: Queue, Status, Priority, Assignee, Author, Type, Resolution, Updated, Created, Due. Operators: : (exact match), >, <, >=, <= (numeric/dates). Examples: 'Status: Open', 'Assignee: me() AND Priority: Critical', '(Assignee: me() OR Author: me()) AND NOT Status: Closed', 'Updated: >today()-7d', 'Queue: CP OR BB AND NOT Status: Closed', 'Resolution: empty()'. IMPORTANT: Cannot be used together with 'filter' - use either filter or query, not both."`
}

// listTransitionsInputDTO is the input for tracker_issue_transitions_list tool.
type listTransitionsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// listQueuesInputDTO is the input for tracker_queues_list tool.
type listQueuesInputDTO struct {
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'projects' (project information), 'components' (queue components), 'versions' (queue versions), 'types' (issue types), 'team' (team members), 'workflows' (workflow configurations), 'all' (all additional fields). Can be combined: 'projects,team'. Example: 'all'"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of queues per page. Valid range: 1-50 (default: 50). Use for pagination when result set exceeds 50 queues."`
	Page    int    `json:"page,omitempty" jsonschema:"Page number for pagination (1-based, default: 1). Use with per_page to navigate through large result sets."`
}

// listCommentsInputDTO is the input for tracker_issue_comments_list tool.
type listCommentsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'attachments' (attached files metadata), 'html' (comment HTML markup), 'all' (all additional fields). Example: 'attachments,html'"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of comments per page. Valid range: 1-50 (default: 50). Use for pagination when issue has many comments."`
	ID      string `json:"id,omitempty" jsonschema:"Comment ID (string) after which the requested page will begin (for pagination). Use with per_page to navigate through comments chronologically. Example: '12345' (numeric ID as string)"`
}

// listAttachmentsInputDTO is the input for tracker_issue_attachments_list tool.
type listAttachmentsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// getAttachmentInputDTO is the input for tracker_issue_attachment_get tool.
type getAttachmentInputDTO struct {
	IssueID      string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	AttachmentID string `json:"attachment_id" jsonschema:"Attachment ID as string. Example: '4159',required"`
	FileName     string `json:"file_name" jsonschema:"Attachment file name including extension. Example: 'attachment.txt',required"`
}

// getAttachmentPreviewInputDTO is the input for tracker_issue_attachment_preview_get tool.
type getAttachmentPreviewInputDTO struct {
	IssueID      string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	AttachmentID string `json:"attachment_id" jsonschema:"Attachment ID as string. Example: '4159',required"`
}

// getQueueInputDTO is the input for tracker_queue_get tool.
type getQueueInputDTO struct {
	QueueID string `json:"queue_id_or_key" jsonschema:"Queue ID or key (e.g., MYQUEUE),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'projects' (project information), 'components' (queue components), 'versions' (queue versions), 'types' (issue types), 'team' (team members), 'workflows' (workflow configurations), 'all' (all additional fields). Example: 'all'"`
}

// getCurrentUserInputDTO is the input for tracker_user_current tool.
type getCurrentUserInputDTO struct {
	// No input required
}

// listUsersInputDTO is the input for tracker_users_list tool.
type listUsersInputDTO struct {
	PerPage int `json:"per_page,omitempty" jsonschema:"Number of users per page. Valid range: 1-50 (default: 50). Use for pagination when organization has many users."`
	Page    int `json:"page,omitempty" jsonschema:"Page number for pagination (1-based, default: 1). Use with per_page to navigate through user list."`
}

// getUserInputDTO is the input for tracker_user_get tool.
type getUserInputDTO struct {
	UserID string `json:"user_id" jsonschema:"User login or ID as string. Accepts either username/login (e.g., 'user.login') or numeric ID as string (e.g., '8000000000000015'),required"`
}

// listLinksInputDTO is the input for tracker_issue_links_list tool.
type listLinksInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// getChangelogInputDTO is the input for tracker_issue_changelog tool.
type getChangelogInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of changelog entries per page. Valid range: 1-50 (default: 50). Use for pagination when issue has extensive history (>50 changes)."`
}

// listProjectCommentsInputDTO is the input for tracker_project_comments_list tool.
type listProjectCommentsInputDTO struct {
	ProjectID string `json:"project_id" jsonschema:"Project ID as string. Obtained from issue.project.primary.id or project list. Example: '114' (numeric ID as string),required"`
	Expand    string `json:"expand,omitempty" jsonschema:"Additional fields to include in response. Possible values: 'all' (all additional fields), 'html' (comment HTML markup), 'attachments' (attached files metadata), 'reactions' (user reactions). Can be combined: 'html,attachments'. Example: 'all'"`
}

// Output DTOs for tracker tools.

// issueOutputDTO represents a Tracker issue.
type issueOutputDTO struct {
	Self            string             `json:"self"`
	ID              string             `json:"id"`
	Key             string             `json:"key"`
	Version         int                `json:"version"`
	Summary         string             `json:"summary"`
	Description     string             `json:"description,omitempty"`
	StatusStartTime string             `json:"status_start_time,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
	UpdatedAt       string             `json:"updated_at,omitempty"`
	ResolvedAt      string             `json:"resolved_at,omitempty"`
	Status          *statusOutputDTO   `json:"status,omitempty"`
	Type            *typeOutputDTO     `json:"type,omitempty"`
	Priority        *priorityOutputDTO `json:"priority,omitempty"`
	Queue           *queueOutputDTO    `json:"queue,omitempty"`
	Assignee        *userOutputDTO     `json:"assignee,omitempty"`
	CreatedBy       *userOutputDTO     `json:"created_by,omitempty"`
	UpdatedBy       *userOutputDTO     `json:"updated_by,omitempty"`
	Votes           int                `json:"votes,omitempty"`
	Favorite        bool               `json:"favorite,omitempty"`
}

// statusOutputDTO represents an issue status.
type statusOutputDTO struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// typeOutputDTO represents an issue type.
type typeOutputDTO struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// priorityOutputDTO represents an issue priority.
type priorityOutputDTO struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// queueOutputDTO represents a Tracker queue.
type queueOutputDTO struct {
	Self           string         `json:"self"`
	ID             string         `json:"id"`
	Key            string         `json:"key"`
	Display        string         `json:"display,omitempty"`
	Name           string         `json:"name,omitempty"`
	Version        int            `json:"version,omitempty"`
	Lead           *userOutputDTO `json:"lead,omitempty"`
	AssignAuto     bool           `json:"assign_auto,omitempty"`
	AllowExternals bool           `json:"allow_externals,omitempty"`
	DenyVoting     bool           `json:"deny_voting,omitempty"`
}

// userOutputDTO represents a Tracker user.
type userOutputDTO struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	UID         string `json:"uid,omitempty"`
	Login       string `json:"login,omitempty"`
	Display     string `json:"display,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	CloudUID    string `json:"cloud_uid,omitempty"`
	PassportUID string `json:"passport_uid,omitempty"`
}

// transitionOutputDTO represents an available issue transition.
type transitionOutputDTO struct {
	ID      string           `json:"id"`
	Display string           `json:"display"`
	Self    string           `json:"self"`
	To      *statusOutputDTO `json:"to,omitempty"`
}

// commentOutputDTO represents an issue comment.
type commentOutputDTO struct {
	ID        string         `json:"id"`
	LongID    string         `json:"long_id"`
	Self      string         `json:"self"`
	Text      string         `json:"text"`
	Version   int            `json:"version"`
	Type      string         `json:"type,omitempty"`
	Transport string         `json:"transport,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	CreatedBy *userOutputDTO `json:"created_by,omitempty"`
	UpdatedBy *userOutputDTO `json:"updated_by,omitempty"`
}

// searchIssuesOutputDTO is the output for tracker_issue_search tool.
type searchIssuesOutputDTO struct {
	Issues      []issueOutputDTO `json:"issues"`
	TotalCount  int              `json:"total_count"`
	TotalPages  int              `json:"total_pages"`
	ScrollID    string           `json:"scroll_id,omitempty"`
	ScrollToken string           `json:"scroll_token,omitempty"`
	NextLink    string           `json:"next_link,omitempty"`
}

// countIssuesOutputDTO is the output for tracker_issue_count tool.
type countIssuesOutputDTO struct {
	Count int `json:"count"`
}

// transitionsListOutputDTO is the output for tracker_issue_transitions_list tool.
type transitionsListOutputDTO struct {
	Transitions []transitionOutputDTO `json:"transitions"`
}

// queuesListOutputDTO is the output for tracker_queues_list tool.
type queuesListOutputDTO struct {
	Queues     []queueOutputDTO `json:"queues"`
	TotalCount int              `json:"total_count"`
	TotalPages int              `json:"total_pages"`
}

// commentsListOutputDTO is the output for tracker_issue_comments_list tool.
type commentsListOutputDTO struct {
	Comments []commentOutputDTO `json:"comments"`
	NextLink string             `json:"next_link,omitempty"`
}

// attachmentOutputDTO represents an issue attachment.
type attachmentOutputDTO struct {
	ID           string                       `json:"id"`
	Name         string                       `json:"name"`
	ContentURL   string                       `json:"content_url"`
	ThumbnailURL string                       `json:"thumbnail_url,omitempty"`
	Mimetype     string                       `json:"mimetype,omitempty"`
	Size         int64                        `json:"size"`
	CreatedAt    string                       `json:"created_at,omitempty"`
	CreatedBy    *userOutputDTO               `json:"created_by,omitempty"`
	Metadata     *attachmentMetadataOutputDTO `json:"metadata,omitempty"`
}

// attachmentMetadataOutputDTO represents attachment metadata.
type attachmentMetadataOutputDTO struct {
	Size string `json:"size,omitempty"`
}

// attachmentsListOutputDTO is the output for tracker_issue_attachments_list tool.
type attachmentsListOutputDTO struct {
	Attachments []attachmentOutputDTO `json:"attachments"`
}

// attachmentContentOutputDTO represents downloaded attachment content.
type attachmentContentOutputDTO struct {
	FileName      string `json:"file_name,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ContentBase64 string `json:"content_base64"`
	Size          int64  `json:"size"`
}

// queueDetailOutputDTO represents a detailed queue response.
type queueDetailOutputDTO struct {
	Self            string             `json:"self"`
	ID              string             `json:"id"`
	Key             string             `json:"key"`
	Display         string             `json:"display,omitempty"`
	Name            string             `json:"name,omitempty"`
	Description     string             `json:"description,omitempty"`
	Version         int                `json:"version,omitempty"`
	Lead            *userOutputDTO     `json:"lead,omitempty"`
	AssignAuto      bool               `json:"assign_auto,omitempty"`
	AllowExternals  bool               `json:"allow_externals,omitempty"`
	DenyVoting      bool               `json:"deny_voting,omitempty"`
	DefaultType     *typeOutputDTO     `json:"default_type,omitempty"`
	DefaultPriority *priorityOutputDTO `json:"default_priority,omitempty"`
}

// userDetailOutputDTO represents a detailed user response.
type userDetailOutputDTO struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	UID         string `json:"uid,omitempty"`
	TrackerUID  string `json:"tracker_uid,omitempty"`
	Login       string `json:"login,omitempty"`
	Display     string `json:"display,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	CloudUID    string `json:"cloud_uid,omitempty"`
	PassportUID string `json:"passport_uid,omitempty"`
	HasLicense  bool   `json:"has_license,omitempty"`
	Dismissed   bool   `json:"dismissed,omitempty"`
	External    bool   `json:"external,omitempty"`
}

// usersListOutputDTO is the output for tracker_users_list tool.
type usersListOutputDTO struct {
	Users      []userDetailOutputDTO `json:"users"`
	TotalCount int                   `json:"total_count,omitempty"`
	TotalPages int                   `json:"total_pages,omitempty"`
}

// linkTypeOutputDTO represents a link type.
type linkTypeOutputDTO struct {
	ID      string `json:"id"`
	Inward  string `json:"inward,omitempty"`
	Outward string `json:"outward,omitempty"`
}

// linkedIssueOutputDTO represents a linked issue reference.
type linkedIssueOutputDTO struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display,omitempty"`
}

// linkOutputDTO represents a link between issues.
type linkOutputDTO struct {
	ID        string                `json:"id"`
	Self      string                `json:"self"`
	Type      *linkTypeOutputDTO    `json:"type,omitempty"`
	Direction string                `json:"direction,omitempty"`
	Object    *linkedIssueOutputDTO `json:"object,omitempty"`
	CreatedBy *userOutputDTO        `json:"created_by,omitempty"`
	UpdatedBy *userOutputDTO        `json:"updated_by,omitempty"`
	CreatedAt string                `json:"created_at,omitempty"`
	UpdatedAt string                `json:"updated_at,omitempty"`
}

// linksListOutputDTO is the output for tracker_issue_links_list tool.
type linksListOutputDTO struct {
	Links []linkOutputDTO `json:"links"`
}

// changelogFieldOutputDTO represents a single field change.
type changelogFieldOutputDTO struct {
	Field string `json:"field"`
	From  any    `json:"from,omitempty"`
	To    any    `json:"to,omitempty"`
}

// changelogEntryOutputDTO represents a single changelog entry.
type changelogEntryOutputDTO struct {
	ID        string                    `json:"id"`
	Self      string                    `json:"self"`
	Issue     *linkedIssueOutputDTO     `json:"issue,omitempty"`
	UpdatedAt string                    `json:"updated_at,omitempty"`
	UpdatedBy *userOutputDTO            `json:"updated_by,omitempty"`
	Type      string                    `json:"type,omitempty"`
	Transport string                    `json:"transport,omitempty"`
	Fields    []changelogFieldOutputDTO `json:"fields,omitempty"`
}

// changelogOutputDTO is the output for tracker_issue_changelog tool.
type changelogOutputDTO struct {
	Entries []changelogEntryOutputDTO `json:"entries"`
}

// projectCommentOutputDTO represents a project comment.
type projectCommentOutputDTO struct {
	ID        string         `json:"id"`
	LongID    string         `json:"long_id,omitempty"`
	Self      string         `json:"self"`
	Text      string         `json:"text,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	CreatedBy *userOutputDTO `json:"created_by,omitempty"`
	UpdatedBy *userOutputDTO `json:"updated_by,omitempty"`
}

// projectCommentsListOutputDTO is the output for tracker_project_comments_list tool.
type projectCommentsListOutputDTO struct {
	Comments []projectCommentOutputDTO `json:"comments"`
}
