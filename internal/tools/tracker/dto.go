//nolint:lll // JSON schema descriptions for LLM tool inputs require detailed inline documentation
package tracker

// Input DTOs for tracker tools.

// getIssueInputDTO is the input for tracker_issue_get tool.
type getIssueInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: attachments (attached files)"`
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
	Filter map[string]string `json:"filter,omitempty" jsonschema:"Field-based filter object (e.g., {\"queue\": \"JUNE\", \"assignee\": \"empty()\"})"`
	Query  string            `json:"query,omitempty" jsonschema:"Query language filter string (yandex tracker syntax)"`
}

// listTransitionsInputDTO is the input for tracker_issue_transitions_list tool.
type listTransitionsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// listQueuesInputDTO is the input for tracker_queues_list tool.
type listQueuesInputDTO struct {
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: projects, components, versions, types, team, workflows"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of queues per page (default: 50)"`
	Page    int    `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
}

// listCommentsInputDTO is the input for tracker_issue_comments_list tool.
type listCommentsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: attachments (attached files), html (comment HTML markup), all (all additional fields)"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of comments per page (default: 50)"`
	ID      string `json:"id,omitempty" jsonschema:"Comment id value after which the requested page will begin (for pagination)"`
}

// Write tool input DTOs.

// createIssueInputDTO is the input for tracker_issue_create tool.
type createIssueInputDTO struct {
	Queue         string   `json:"queue" jsonschema:"Queue key (e.g., TEST),required"`
	Summary       string   `json:"summary" jsonschema:"Issue summary,required"`
	Description   string   `json:"description,omitempty" jsonschema:"Issue description"`
	Type          string   `json:"type,omitempty" jsonschema:"Issue type key (e.g., bug, task, story)"`
	Priority      string   `json:"priority,omitempty" jsonschema:"Priority key (e.g., critical, normal, low)"`
	Assignee      string   `json:"assignee,omitempty" jsonschema:"Assignee login"`
	Tags          []string `json:"tags,omitempty" jsonschema:"Issue tags"`
	Parent        string   `json:"parent,omitempty" jsonschema:"Parent issue key (e.g., TEST-1)"`
	AttachmentIDs []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs (strings) to link"`
	Sprint        []string `json:"sprint,omitempty" jsonschema:"Sprint IDs (strings) to add issue to"`
}

// updateIssueInputDTO is the input for tracker_issue_update tool.
type updateIssueInputDTO struct {
	IssueID             string   `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Summary             string   `json:"summary,omitempty" jsonschema:"Issue summary"`
	Description         string   `json:"description,omitempty" jsonschema:"Issue description"`
	Type                string   `json:"type,omitempty" jsonschema:"Issue type key"`
	Priority            string   `json:"priority,omitempty" jsonschema:"Priority key"`
	Assignee            string   `json:"assignee,omitempty" jsonschema:"Assignee login"`
	Version             int      `json:"version,omitempty" jsonschema:"Issue version for optimistic locking"`
	ProjectPrimary      int      `json:"project_primary,omitempty" jsonschema:"Primary project ID"`
	ProjectSecondaryAdd []int    `json:"project_secondary_add,omitempty" jsonschema:"Secondary project IDs to add"`
	Sprint              []string `json:"sprint,omitempty" jsonschema:"Sprint IDs (strings) or keys to assign"`
}

// executeTransitionInputDTO is the input for tracker_issue_transition_execute tool.
type executeTransitionInputDTO struct {
	IssueID      string         `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	TransitionID string         `json:"transition_id" jsonschema:"Transition ID (string),required"`
	Comment      string         `json:"comment,omitempty" jsonschema:"Comment to add during transition"`
	Fields       map[string]any `json:"fields,omitempty" jsonschema:"Additional fields to set during transition"`
}

// addCommentInputDTO is the input for tracker_issue_comment_add tool.
type addCommentInputDTO struct {
	IssueID           string   `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Text              string   `json:"text" jsonschema:"Comment text,required"`
	AttachmentIDs     []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs (strings) to link"`
	MarkupType        string   `json:"markup_type,omitempty" jsonschema:"Text markup type. Use md for YFM markup"`
	Summonees         []string `json:"summonees,omitempty" jsonschema:"User logins to summon"`
	MaillistSummonees []string `json:"maillist_summonees,omitempty" jsonschema:"Mailing list addresses to summon"`
	IsAddToFollowers  bool     `json:"is_add_to_followers,omitempty" jsonschema:"Add summoned users to followers"`
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

// updateCommentInputDTO is the input for tracker_issue_comment_update tool.
type updateCommentInputDTO struct {
	IssueID           string   `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	CommentID         string   `json:"comment_id" jsonschema:"Comment ID (string),required"`
	Text              string   `json:"text" jsonschema:"Comment text,required"`
	AttachmentIDs     []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs (strings) to link"`
	MarkupType        string   `json:"markup_type,omitempty" jsonschema:"Text markup type. Use md for YFM markup"`
	Summonees         []string `json:"summonees,omitempty" jsonschema:"User logins to summon"`
	MaillistSummonees []string `json:"maillist_summonees,omitempty" jsonschema:"Mailing list addresses to summon"`
}

// deleteCommentInputDTO is the input for tracker_issue_comment_delete tool.
type deleteCommentInputDTO struct {
	IssueID   string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	CommentID string `json:"comment_id" jsonschema:"Comment ID (string),required"`
}

// listAttachmentsInputDTO is the input for tracker_issue_attachments_list tool.
type listAttachmentsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// deleteAttachmentInputDTO is the input for tracker_issue_attachment_delete tool.
type deleteAttachmentInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	FileID  string `json:"file_id" jsonschema:"Attachment file ID (string),required"`
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

// deleteCommentOutputDTO is the output for tracker_issue_comment_delete tool.
type deleteCommentOutputDTO struct {
	Success bool `json:"success"`
}

// deleteAttachmentOutputDTO is the output for tracker_issue_attachment_delete tool.
type deleteAttachmentOutputDTO struct {
	Success bool `json:"success"`
}

// getQueueInputDTO is the input for tracker_queue_get tool.
type getQueueInputDTO struct {
	QueueID string `json:"queue_id_or_key" jsonschema:"Queue ID or key (e.g., MYQUEUE),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: projects, components, versions, types, team, workflows, all"`
}

// createQueueInputDTO is the input for tracker_queue_create tool.
type createQueueInputDTO struct {
	Key             string `json:"key" jsonschema:"Queue key (e.g., MYQUEUE),required"`
	Name            string `json:"name" jsonschema:"Queue name,required"`
	Lead            string `json:"lead" jsonschema:"Queue lead login or user ID,required"`
	DefaultType     string `json:"default_type" jsonschema:"Default issue type key or ID,required"`
	DefaultPriority string `json:"default_priority" jsonschema:"Default priority key or ID,required"`
}

// deleteQueueInputDTO is the input for tracker_queue_delete tool.
type deleteQueueInputDTO struct {
	QueueID string `json:"queue_id_or_key" jsonschema:"Queue ID or key (e.g., MYQUEUE),required"`
}

// restoreQueueInputDTO is the input for tracker_queue_restore tool.
type restoreQueueInputDTO struct {
	QueueID string `json:"queue_id_or_key" jsonschema:"Queue ID or key (e.g., MYQUEUE),required"`
}

// getCurrentUserInputDTO is the input for tracker_user_current tool.
type getCurrentUserInputDTO struct {
	// No input required
}

// listUsersInputDTO is the input for tracker_users_list tool.
type listUsersInputDTO struct {
	PerPage int `json:"per_page,omitempty" jsonschema:"Number of users per page (default: 50)"`
	Page    int `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
}

// getUserInputDTO is the input for tracker_user_get tool.
type getUserInputDTO struct {
	UserID string `json:"user_id" jsonschema:"User login or ID,required"`
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

// deleteQueueOutputDTO is the output for tracker_queue_delete tool.
type deleteQueueOutputDTO struct {
	Success bool `json:"success"`
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

// listLinksInputDTO is the input for tracker_issue_links_list tool.
type listLinksInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// createLinkInputDTO is the input for tracker_issue_link_create tool.
type createLinkInputDTO struct {
	IssueID      string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Relationship string `json:"relationship" jsonschema:"Link type ID (e.g., relates, depends, duplicates),required"`
	TargetIssue  string `json:"target_issue" jsonschema:"Target issue ID or key to link to,required"`
}

// deleteLinkInputDTO is the input for tracker_issue_link_delete tool.
type deleteLinkInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	LinkID  string `json:"link_id" jsonschema:"Link ID (string) to delete,required"`
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

// deleteLinkOutputDTO is the output for tracker_issue_link_delete tool.
type deleteLinkOutputDTO struct {
	Success bool `json:"success"`
}

// getChangelogInputDTO is the input for tracker_issue_changelog tool.
type getChangelogInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of changelog entries per page (default: 50)"`
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

// moveIssueInputDTO is the input for tracker_issue_move tool.
type moveIssueInputDTO struct {
	IssueID       string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Queue         string `json:"queue" jsonschema:"Target queue key (e.g., NEWQUEUE),required"`
	InitialStatus bool   `json:"initial_status,omitempty" jsonschema:"Reset issue status to initial value when moving"`
}

// listProjectCommentsInputDTO is the input for tracker_project_comments_list tool.
type listProjectCommentsInputDTO struct {
	ProjectID string `json:"project_id" jsonschema:"Project ID (string),required"`
	Expand    string `json:"expand,omitempty" jsonschema:"Additional fields to include. Possible values: all, html, attachments, reactions"`
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
