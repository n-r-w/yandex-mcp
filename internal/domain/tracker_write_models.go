package domain

// TrackerIssueCreateRequest represents a request to create a new issue.
type TrackerIssueCreateRequest struct {
	Queue         string
	Summary       string
	Description   string
	Type          string
	Priority      string
	Assignee      string
	Tags          []string
	Parent        string
	AttachmentIDs []string
	Sprint        []string
}

// TrackerIssueUpdateRequest represents a request to update an existing issue.
type TrackerIssueUpdateRequest struct {
	IssueID             string
	Summary             string
	Description         string
	Type                string
	Priority            string
	Assignee            string
	Version             int
	ProjectPrimary      int
	ProjectSecondaryAdd []int
	Sprint              []string
}

// TrackerTransitionExecuteRequest represents a request to execute an issue transition.
type TrackerTransitionExecuteRequest struct {
	IssueID      string
	TransitionID string
	Comment      string
	Fields       map[string]any
}

// TrackerCommentAddRequest represents a request to add a comment to an issue.
type TrackerCommentAddRequest struct {
	IssueID           string
	Text              string
	AttachmentIDs     []string
	MarkupType        string
	Summonees         []string
	MaillistSummonees []string
	IsAddToFollowers  bool
}

// TrackerIssueCreateResponse represents the response from issue creation.
type TrackerIssueCreateResponse struct {
	Issue TrackerIssue
}

// TrackerIssueUpdateResponse represents the response from issue update.
type TrackerIssueUpdateResponse struct {
	Issue TrackerIssue
}

// TrackerTransitionExecuteResponse represents the response from transition execution.
type TrackerTransitionExecuteResponse struct {
	Transitions []TrackerTransition
}

// TrackerCommentAddResponse represents the response from comment addition.
type TrackerCommentAddResponse struct {
	Comment TrackerComment
}

// TrackerCommentUpdateRequest represents a request to update an existing comment.
type TrackerCommentUpdateRequest struct {
	IssueID           string
	CommentID         string
	Text              string
	AttachmentIDs     []string
	MarkupType        string
	Summonees         []string
	MaillistSummonees []string
}

// TrackerCommentUpdateResponse represents the response from comment update.
type TrackerCommentUpdateResponse struct {
	Comment TrackerComment
}

// TrackerCommentDeleteRequest represents a request to delete a comment.
type TrackerCommentDeleteRequest struct {
	IssueID   string
	CommentID string
}

// TrackerAttachmentDeleteRequest represents a request to delete an attachment.
type TrackerAttachmentDeleteRequest struct {
	IssueID string
	FileID  string
}

// TrackerQueueCreateRequest represents a request to create a new queue.
type TrackerQueueCreateRequest struct {
	Key             string
	Name            string
	Lead            string
	DefaultType     string
	DefaultPriority string
}

// TrackerQueueCreateResponse represents the response from queue creation.
type TrackerQueueCreateResponse struct {
	Queue TrackerQueueDetail
}

// TrackerQueueDeleteRequest represents a request to delete a queue.
type TrackerQueueDeleteRequest struct {
	QueueID string
}

// TrackerQueueRestoreRequest represents a request to restore a deleted queue.
type TrackerQueueRestoreRequest struct {
	QueueID string
}

// TrackerQueueRestoreResponse represents the response from queue restoration.
type TrackerQueueRestoreResponse struct {
	Queue TrackerQueueDetail
}

// TrackerLinkCreateRequest represents a request to create a link between issues.
type TrackerLinkCreateRequest struct {
	IssueID      string
	Relationship string
	TargetIssue  string
}

// TrackerLinkCreateResponse represents the response from link creation.
type TrackerLinkCreateResponse struct {
	Link TrackerLink
}

// TrackerLinkDeleteRequest represents a request to delete a link.
type TrackerLinkDeleteRequest struct {
	IssueID string
	LinkID  string
}

// TrackerIssueMoveRequest represents a request to move an issue to another queue.
type TrackerIssueMoveRequest struct {
	IssueID       string
	Queue         string
	InitialStatus bool
}

// TrackerIssueMoveResponse represents the response from issue move.
type TrackerIssueMoveResponse struct {
	Issue TrackerIssue
}
