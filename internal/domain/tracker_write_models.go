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
	IssueID     string
	Summary     string
	Description string
	Type        string
	Priority    string
	Assignee    string
	Version     int
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
