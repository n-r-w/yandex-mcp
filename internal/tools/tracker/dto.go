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
	// Filter is a field-based filter object (e.g., {\"queue\": \"TREK\", \"assignee\": \"empty()\"}).
	Filter map[string]any `json:"filter,omitempty" jsonschema:"Field-based filter object with key-value pairs"`
	// Query is a query language filter string.
	Query string `json:"query,omitempty" jsonschema:"Query language filter string (yandex tracker query syntax). E.g.: Queue:\"TEST\" Status:\"Open\",\"In progress\" (Assignee:me() OR Author:me()) Resolution:empty() Updated:>today()-1w"`
	// Order specifies sorting field and direction (e.g. +updated or -created).
	Order string `json:"order,omitempty" jsonschema:"Issue sorting direction and field. Format: [+/-]<field_key>. NOTE: Only used together with filter parameter, not with query"`
	// Expand specifies additional fields to include.
	Expand string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: transitions (workflow transitions between statuses), attachments (attached files)"`
	// PerPage is the number of results per page (standard pagination).
	PerPage int `json:"per_page,omitempty" jsonschema:"Number of results per page (default: 50). Used for standard pagination (<10,000 results)"`
	// Page is the page number (standard pagination).
	Page int `json:"page,omitempty" jsonschema:"Page number (default: 1). Used for standard pagination (<10,000 results)"`
	// ScrollType is the scroll type: sorted or unsorted.
	ScrollType string `json:"scroll_type,omitempty" jsonschema:"Scroll type for large result sets (>10,000). Possible values: sorted (use sorting specified in request), unsorted (no sorting). NOTE: Used only in first request of scrollable sequence"`
	// PerScroll is the max issues per scroll response (max 1000).
	PerScroll int `json:"per_scroll,omitempty" jsonschema:"Max issues per scroll response (max: 1000, default: 100). Used only in first scroll request"`
	// ScrollTTLMillis is the scroll context lifetime in milliseconds.
	ScrollTTLMillis int `json:"scroll_ttl_millis,omitempty" jsonschema:"Scroll context lifetime in milliseconds (default: 60000). Used only in first scroll request"`
	// ScrollID is the scroll page ID for subsequent requests.
	ScrollID string `json:"scroll_id,omitempty" jsonschema:"Scroll page ID for 2nd and subsequent scroll requests"`
}

// countIssuesInputDTO is the input for tracker_issue_count tool.
type countIssuesInputDTO struct {
	Filter map[string]any `json:"filter,omitempty" jsonschema:"Field-based filter object (e.g., {\"queue\": \"JUNE\", \"assignee\": \"empty()\"})"`
	Query  string         `json:"query,omitempty" jsonschema:"Query language filter string (yandex tracker syntax)"`
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
	AttachmentIDs []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs to link"`
	Sprint        []string `json:"sprint,omitempty" jsonschema:"Sprint IDs to add issue to"`
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
	Sprint              []string `json:"sprint,omitempty" jsonschema:"Sprint IDs or keys to assign"`
}

// executeTransitionInputDTO is the input for tracker_issue_transition_execute tool.
type executeTransitionInputDTO struct {
	IssueID      string         `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	TransitionID string         `json:"transition_id" jsonschema:"Transition ID,required"`
	Comment      string         `json:"comment,omitempty" jsonschema:"Comment to add during transition"`
	Fields       map[string]any `json:"fields,omitempty" jsonschema:"Additional fields to set during transition"`
}

// addCommentInputDTO is the input for tracker_issue_comment_add tool.
type addCommentInputDTO struct {
	IssueID           string   `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Text              string   `json:"text" jsonschema:"Comment text,required"`
	AttachmentIDs     []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs to link"`
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
	CommentID         string   `json:"comment_id" jsonschema:"Comment ID,required"`
	Text              string   `json:"text" jsonschema:"Comment text,required"`
	AttachmentIDs     []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs to link"`
	MarkupType        string   `json:"markup_type,omitempty" jsonschema:"Text markup type. Use md for YFM markup"`
	Summonees         []string `json:"summonees,omitempty" jsonschema:"User logins to summon"`
	MaillistSummonees []string `json:"maillist_summonees,omitempty" jsonschema:"Mailing list addresses to summon"`
}

// deleteCommentInputDTO is the input for tracker_issue_comment_delete tool.
type deleteCommentInputDTO struct {
	IssueID   string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	CommentID string `json:"comment_id" jsonschema:"Comment ID,required"`
}

// listAttachmentsInputDTO is the input for tracker_issue_attachments_list tool.
type listAttachmentsInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// deleteAttachmentInputDTO is the input for tracker_issue_attachment_delete tool.
type deleteAttachmentInputDTO struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	FileID  string `json:"file_id" jsonschema:"Attachment file ID,required"`
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
	LinkID  string `json:"link_id" jsonschema:"Link ID to delete,required"`
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
	ProjectID string `json:"project_id" jsonschema:"Project ID or short ID,required"`
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
