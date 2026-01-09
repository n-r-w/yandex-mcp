package tracker

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Issue is a Yandex Tracker issue.
type Issue struct {
	Self            string  `json:"self"`
	ID              string  `json:"id"`
	Key             string  `json:"key"`
	Version         int     `json:"version"`
	Summary         string  `json:"summary"`
	Description     string  `json:"description,omitempty"`
	StatusStartTime string  `json:"statusStartTime,omitempty"`
	CreatedAt       string  `json:"createdAt,omitempty"`
	UpdatedAt       string  `json:"updatedAt,omitempty"`
	ResolvedAt      string  `json:"resolvedAt,omitempty"`
	Status          *Status `json:"status,omitempty"`
	Type            *Type   `json:"type,omitempty"`
	Priority        *Prio   `json:"priority,omitempty"`
	Queue           *Queue  `json:"queue,omitempty"`
	Assignee        *User   `json:"assignee,omitempty"`
	CreatedBy       *User   `json:"createdBy,omitempty"`
	UpdatedBy       *User   `json:"updatedBy,omitempty"`
	Votes           int     `json:"votes,omitempty"`
	Favorite        bool    `json:"favorite,omitempty"`
}

// Status represents an issue status.
type Status struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// Type represents an issue type.
type Type struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// Prio represents an issue priority.
type Prio struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

// Queue represents a Tracker queue.
type Queue struct {
	Self           string `json:"self"`
	ID             string `json:"id"`
	Key            string `json:"key"`
	Display        string `json:"display,omitempty"`
	Name           string `json:"name,omitempty"`
	Version        int    `json:"version,omitempty"`
	Lead           *User  `json:"lead,omitempty"`
	AssignAuto     bool   `json:"assignAuto,omitempty"`
	AllowExternals bool   `json:"allowExternals,omitempty"`
	DenyVoting     bool   `json:"denyVoting,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Queue to handle numeric and string IDs.
func (q *Queue) UnmarshalJSON(data []byte) error {
	type QueueAlias Queue

	alias := &struct {
		ID any `json:"id"` // Queue ID can be either string or number
		*QueueAlias
	}{
		ID:         nil,
		QueueAlias: (*QueueAlias)(q),
	}

	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	// Convert ID to string regardless of original type
	if alias.ID != nil {
		switch v := alias.ID.(type) {
		case float64:
			q.ID = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			q.ID = v
		default:
			return fmt.Errorf("unsupported queue ID type: %T", v)
		}
	}

	return nil
}

// User represents a Tracker user.
type User struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	UID         int64  `json:"uid,omitempty"`
	Login       string `json:"login,omitempty"`
	Display     string `json:"display,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty"`
	CloudUID    string `json:"cloudUid,omitempty"`
	PassportUID int64  `json:"passportUid,omitempty"`
}

// Transition represents an available issue transition.
type Transition struct {
	ID      string  `json:"id"`
	Display string  `json:"display"`
	Self    string  `json:"self"`
	To      *Status `json:"to"`
}

// Comment represents an issue comment.
type Comment struct {
	ID        int64  `json:"id"`
	LongID    string `json:"longId"`
	Self      string `json:"self"`
	Text      string `json:"text"`
	Version   int    `json:"version"`
	Type      string `json:"type,omitempty"`
	Transport string `json:"transport,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	CreatedBy *User  `json:"createdBy,omitempty"`
	UpdatedBy *User  `json:"updatedBy,omitempty"`
}

// SearchIssuesOpts contains options for searching issues.
type SearchIssuesOpts struct {
	// Filter is a field-based filter object.
	Filter map[string]any
	// Query is a query language filter string.
	Query string
	// Order specifies sorting direction and field (e.g., "+updated", "-created").
	Order string
	// Expand specifies additional fields to include (transitions, attachments).
	Expand string
	// PerPage specifies the number of results per page (standard pagination).
	PerPage int
	// Page specifies the page number (standard pagination).
	Page int
	// ScrollType specifies scrolling type: "sorted" or "unsorted".
	ScrollType string
	// PerScroll specifies max issues per response in scroll mode (max 1000).
	PerScroll int
	// ScrollTTLMillis specifies scroll context lifetime in milliseconds.
	ScrollTTLMillis int
	// ScrollID specifies the scroll page ID for subsequent requests.
	ScrollID string
}

// SearchIssuesResult contains the result of a search operation.
type SearchIssuesResult struct {
	Issues      []Issue
	TotalCount  int
	TotalPages  int
	ScrollID    string
	ScrollToken string
	NextLink    string
}

// CountIssuesOpts contains options for counting issues.
type CountIssuesOpts struct {
	// Filter is a field-based filter object.
	Filter map[string]any
	// Query is a query language filter string.
	Query string
}

// ListQueuesOpts contains options for listing queues.
type ListQueuesOpts struct {
	// Expand specifies additional fields (projects, components, versions, types, team, workflows).
	Expand string
	// PerPage specifies the number of queues per page.
	PerPage int
	// Page specifies the page number.
	Page int
}

// ListQueuesResult contains the result of listing queues.
type ListQueuesResult struct {
	Queues     []Queue
	TotalCount int
	TotalPages int
}

// ListCommentsOpts contains options for listing comments.
type ListCommentsOpts struct {
	// Expand specifies additional fields (attachments, html, all).
	Expand string
	// PerPage specifies the number of comments per page.
	PerPage int
	// ID specifies the comment ID after which the requested page begins.
	ID int64
}

// ListCommentsResult contains the result of listing comments.
type ListCommentsResult struct {
	Comments []Comment
	NextLink string
}

// GetIssueOpts contains options for getting an issue.
type GetIssueOpts struct {
	// Expand specifies additional fields to include (attachments).
	Expand string
}

// searchRequest represents the request body for issue search.
type searchRequest struct {
	Filter map[string]any `json:"filter,omitempty"`
	Query  string         `json:"query,omitempty"`
	Order  string         `json:"order,omitempty"`
}

// countRequest represents the request body for issue count.
type countRequest struct {
	Filter map[string]any `json:"filter,omitempty"`
	Query  string         `json:"query,omitempty"`
}

// errorResponse represents the Tracker API error format.
type errorResponse struct {
	Errors        []string `json:"errors,omitempty"`
	ErrorMessages []string `json:"errorMessages,omitempty"`
	StatusCode    int      `json:"statusCode,omitempty"`
}

// Write operation request DTOs.

// CreateIssueRequest is the request body for issue creation.
type CreateIssueRequest struct {
	Queue         string   `json:"queue"`
	Summary       string   `json:"summary"`
	Description   string   `json:"description,omitempty"`
	Type          string   `json:"type,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	Assignee      string   `json:"assignee,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Parent        string   `json:"parent,omitempty"`
	AttachmentIDs []string `json:"attachmentIds,omitempty"`
	Sprint        []string `json:"sprint,omitempty"`
}

// UpdateIssueRequest is the request body for issue update.
type UpdateIssueRequest struct {
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Assignee    string `json:"assignee,omitempty"`
	Version     int    `json:"version,omitempty"`
}

// ExecuteTransitionRequest is the request body for transition execution.
type ExecuteTransitionRequest struct {
	Comment string         `json:"comment,omitempty"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// AddCommentRequest is the request body for adding a comment.
type AddCommentRequest struct {
	Text              string   `json:"text"`
	AttachmentIDs     []string `json:"attachmentIds,omitempty"`
	MarkupType        string   `json:"markupType,omitempty"`
	Summonees         []string `json:"summonees,omitempty"`
	MaillistSummonees []string `json:"maillistSummonees,omitempty"`
	IsAddToFollowers  *bool    `json:"isAddToFollowers,omitempty"`
}
