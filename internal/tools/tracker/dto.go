//nolint:lll
package tracker

// Input DTOs for tracker tools.

// GetIssueInput is the input for tracker_issue_get tool.
type GetIssueInput struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: attachments (attached files)"`
}

// SearchIssuesInput is the input for tracker_issue_search tool.
type SearchIssuesInput struct {
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

// CountIssuesInput is the input for tracker_issue_count tool.
type CountIssuesInput struct {
	Filter map[string]any `json:"filter,omitempty" jsonschema:"Field-based filter object (e.g., {\"queue\": \"JUNE\", \"assignee\": \"empty()\"})"`
	Query  string         `json:"query,omitempty" jsonschema:"Query language filter string (yandex tracker syntax)"`
}

// ListTransitionsInput is the input for tracker_issue_transitions_list tool.
type ListTransitionsInput struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
}

// ListQueuesInput is the input for tracker_queues_list tool.
type ListQueuesInput struct {
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: projects, components, versions, types, team, workflows"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of queues per page (default: 50)"`
	Page    int    `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
}

// ListCommentsInput is the input for tracker_issue_comments_list tool.
type ListCommentsInput struct {
	IssueID string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Expand  string `json:"expand,omitempty" jsonschema:"Additional fields to include in the response. Possible values: attachments (attached files), html (comment HTML markup), all (all additional fields)"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Number of comments per page (default: 50)"`
	ID      int64  `json:"id,omitempty" jsonschema:"Comment numeric id value after which the requested page will begin (for pagination)"`
}

// Write tool input DTOs.

// CreateIssueInput is the input for tracker_issue_create tool.
type CreateIssueInput struct {
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

// UpdateIssueInput is the input for tracker_issue_update tool.
type UpdateIssueInput struct {
	IssueID     string `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Summary     string `json:"summary,omitempty" jsonschema:"Issue summary"`
	Description string `json:"description,omitempty" jsonschema:"Issue description"`
	Type        string `json:"type,omitempty" jsonschema:"Issue type key"`
	Priority    string `json:"priority,omitempty" jsonschema:"Priority key"`
	Assignee    string `json:"assignee,omitempty" jsonschema:"Assignee login"`
	Version     int    `json:"version,omitempty" jsonschema:"Issue version for optimistic locking"`
}

// ExecuteTransitionInput is the input for tracker_issue_transition_execute tool.
type ExecuteTransitionInput struct {
	IssueID      string         `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	TransitionID string         `json:"transition_id" jsonschema:"Transition ID,required"`
	Comment      string         `json:"comment,omitempty" jsonschema:"Comment to add during transition"`
	Fields       map[string]any `json:"fields,omitempty" jsonschema:"Additional fields to set during transition"`
}

// AddCommentInput is the input for tracker_issue_comment_add tool.
type AddCommentInput struct {
	IssueID           string   `json:"issue_id_or_key" jsonschema:"Issue ID or key (e.g., TEST-1),required"`
	Text              string   `json:"text" jsonschema:"Comment text,required"`
	AttachmentIDs     []string `json:"attachment_ids,omitempty" jsonschema:"Attachment IDs to link"`
	MarkupType        string   `json:"markup_type,omitempty" jsonschema:"Text markup type (plain, wiki, html)"`
	Summonees         []string `json:"summonees,omitempty" jsonschema:"User logins to summon"`
	MaillistSummonees []string `json:"maillist_summonees,omitempty" jsonschema:"Mailing list addresses to summon"`
	IsAddToFollowers  bool     `json:"is_add_to_followers,omitempty" jsonschema:"Add summoned users to followers"`
}

// Output DTOs for tracker tools.

// IssueOutput represents a Tracker issue.
type IssueOutput struct {
	Self            string          `json:"self"`
	ID              string          `json:"id"`
	Key             string          `json:"key"`
	Version         int             `json:"version"`
	Summary         string          `json:"summary"`
	Description     string          `json:"description,omitempty"`
	StatusStartTime string          `json:"status_start_time,omitempty"`
	CreatedAt       string          `json:"created_at,omitempty"`
	UpdatedAt       string          `json:"updated_at,omitempty"`
	ResolvedAt      string          `json:"resolved_at,omitempty"`
	Status          *StatusOutput   `json:"status,omitempty"`
	Type            *TypeOutput     `json:"type,omitempty"`
	Priority        *PriorityOutput `json:"priority,omitempty"`
	Queue           *QueueOutput    `json:"queue,omitempty"`
	Assignee        *UserOutput     `json:"assignee,omitempty"`
	CreatedBy       *UserOutput     `json:"created_by,omitempty"`
	UpdatedBy       *UserOutput     `json:"updated_by,omitempty"`
	Votes           int             `json:"votes,omitempty"`
	Favorite        bool            `json:"favorite,omitempty"`
}

// StatusOutput represents an issue status.
type StatusOutput struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// TypeOutput represents an issue type.
type TypeOutput struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// PriorityOutput represents an issue priority.
type PriorityOutput struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// QueueOutput represents a Tracker queue.
type QueueOutput struct {
	Self           string      `json:"self"`
	ID             string      `json:"id"`
	Key            string      `json:"key"`
	Display        string      `json:"display,omitempty"`
	Name           string      `json:"name,omitempty"`
	Version        int         `json:"version,omitempty"`
	Lead           *UserOutput `json:"lead,omitempty"`
	AssignAuto     bool        `json:"assign_auto,omitempty"`
	AllowExternals bool        `json:"allow_externals,omitempty"`
	DenyVoting     bool        `json:"deny_voting,omitempty"`
}

// UserOutput represents a Tracker user.
type UserOutput struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	UID         int64  `json:"uid,omitempty"`
	Login       string `json:"login,omitempty"`
	Display     string `json:"display,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	CloudUID    string `json:"cloud_uid,omitempty"`
	PassportUID int64  `json:"passport_uid,omitempty"`
}

// TransitionOutput represents an available issue transition.
type TransitionOutput struct {
	ID      string        `json:"id"`
	Display string        `json:"display"`
	Self    string        `json:"self"`
	To      *StatusOutput `json:"to,omitempty"`
}

// CommentOutput represents an issue comment.
type CommentOutput struct {
	ID        int64       `json:"id"`
	LongID    string      `json:"long_id"`
	Self      string      `json:"self"`
	Text      string      `json:"text"`
	Version   int         `json:"version"`
	Type      string      `json:"type,omitempty"`
	Transport string      `json:"transport,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	CreatedBy *UserOutput `json:"created_by,omitempty"`
	UpdatedBy *UserOutput `json:"updated_by,omitempty"`
}

// SearchIssuesOutput is the output for tracker_issue_search tool.
type SearchIssuesOutput struct {
	Issues      []IssueOutput `json:"issues"`
	TotalCount  int           `json:"total_count"`
	TotalPages  int           `json:"total_pages"`
	ScrollID    string        `json:"scroll_id,omitempty"`
	ScrollToken string        `json:"scroll_token,omitempty"`
	NextLink    string        `json:"next_link,omitempty"`
}

// CountIssuesOutput is the output for tracker_issue_count tool.
type CountIssuesOutput struct {
	Count int `json:"count"`
}

// TransitionsListOutput is the output for tracker_issue_transitions_list tool.
type TransitionsListOutput struct {
	Transitions []TransitionOutput `json:"transitions"`
}

// QueuesListOutput is the output for tracker_queues_list tool.
type QueuesListOutput struct {
	Queues     []QueueOutput `json:"queues"`
	TotalCount int           `json:"total_count"`
	TotalPages int           `json:"total_pages"`
}

// CommentsListOutput is the output for tracker_issue_comments_list tool.
type CommentsListOutput struct {
	Comments []CommentOutput `json:"comments"`
	NextLink string          `json:"next_link,omitempty"`
}
