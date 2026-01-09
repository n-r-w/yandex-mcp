package domain

// TrackerIssue represents a Yandex Tracker issue entity.
type TrackerIssue struct {
	Self            string
	ID              string
	Key             string
	Version         int
	Summary         string
	Description     string
	StatusStartTime string
	CreatedAt       string
	UpdatedAt       string
	ResolvedAt      string
	Status          *TrackerStatus
	Type            *TrackerIssueType
	Priority        *TrackerPriority
	Queue           *TrackerQueue
	Assignee        *TrackerUser
	CreatedBy       *TrackerUser
	UpdatedBy       *TrackerUser
	Votes           int
	Favorite        bool
}

// TrackerStatus represents an issue status in Yandex Tracker.
type TrackerStatus struct {
	Self    string
	ID      string
	Key     string
	Display string
}

// TrackerIssueType represents an issue type in Yandex Tracker.
type TrackerIssueType struct {
	Self    string
	ID      string
	Key     string
	Display string
}

// TrackerPriority represents an issue priority in Yandex Tracker.
type TrackerPriority struct {
	Self    string
	ID      string
	Key     string
	Display string
}

// TrackerQueue represents a queue in Yandex Tracker.
type TrackerQueue struct {
	Self           string
	ID             string
	Key            string
	Display        string
	Name           string
	Version        int
	Lead           *TrackerUser
	AssignAuto     bool
	AllowExternals bool
	DenyVoting     bool
}

// TrackerUser represents a user in Yandex Tracker.
type TrackerUser struct {
	Self        string
	ID          string
	UID         int64
	Login       string
	Display     string
	FirstName   string
	LastName    string
	Email       string
	CloudUID    string
	PassportUID int64
}

// TrackerTransition represents a workflow transition for an issue.
type TrackerTransition struct {
	ID      string
	Display string
	Self    string
	To      *TrackerStatus
}

// TrackerComment represents a comment on an issue.
type TrackerComment struct {
	ID        int64
	LongID    string
	Self      string
	Text      string
	Version   int
	Type      string
	Transport string
	CreatedAt string
	UpdatedAt string
	CreatedBy *TrackerUser
	UpdatedBy *TrackerUser
}

// TrackerIssuesPage represents a paginated list of issues.
type TrackerIssuesPage struct {
	Issues      []TrackerIssue
	TotalCount  int
	TotalPages  int
	ScrollID    string
	ScrollToken string
	NextLink    string
}

// TrackerQueuesPage represents a paginated list of queues.
type TrackerQueuesPage struct {
	Queues     []TrackerQueue
	TotalCount int
	TotalPages int
}

// TrackerCommentsPage represents a paginated list of comments.
type TrackerCommentsPage struct {
	Comments []TrackerComment
	NextLink string
}
