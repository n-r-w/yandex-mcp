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

// TrackerQueueDetail represents a detailed queue in Yandex Tracker.
// Extends TrackerQueue with additional fields available when getting queue details.
type TrackerQueueDetail struct {
	Self            string
	ID              string
	Key             string
	Display         string
	Name            string
	Description     string
	Version         int
	Lead            *TrackerUser
	AssignAuto      bool
	AllowExternals  bool
	DenyVoting      bool
	DefaultType     *TrackerIssueType
	DefaultPriority *TrackerPriority
}

// TrackerUser represents a user in Yandex Tracker.
type TrackerUser struct {
	Self        string
	ID          string
	UID         string
	Login       string
	Display     string
	FirstName   string
	LastName    string
	Email       string
	CloudUID    string
	PassportUID string
}

// TrackerUserDetail represents a detailed user in Yandex Tracker.
// Extends TrackerUser with additional fields available when getting user details.
type TrackerUserDetail struct {
	Self        string
	ID          string
	UID         string
	TrackerUID  string
	Login       string
	Display     string
	FirstName   string
	LastName    string
	Email       string
	CloudUID    string
	PassportUID string
	HasLicense  bool
	Dismissed   bool
	External    bool
}

// TrackerUsersPage represents a paginated list of users.
type TrackerUsersPage struct {
	Users      []TrackerUserDetail
	TotalCount int
	TotalPages int
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
	ID        string
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

// TrackerAttachment represents a file attachment in Yandex Tracker.
type TrackerAttachment struct {
	ID           string
	Name         string
	ContentURL   string
	ThumbnailURL string
	Mimetype     string
	Size         int64
	CreatedAt    string
	CreatedBy    *TrackerUser
	Metadata     *TrackerAttachmentMetadata
}

// TrackerAttachmentMetadata holds extra metadata for an attachment.
type TrackerAttachmentMetadata struct {
	Size string
}

// TrackerAttachmentContent represents downloaded attachment content.
type TrackerAttachmentContent struct {
	FileName    string
	ContentType string
	Data        []byte
}

// IAttachmentStream provides streaming access to attachment content.
type IAttachmentStream interface {
	Read(p []byte) (int, error)
	Close() error
}

// TrackerAttachmentStream represents streamed attachment content.
type TrackerAttachmentStream struct {
	FileName    string
	ContentType string
	Stream      IAttachmentStream
}

// TrackerLinkType represents a link type between issues.
type TrackerLinkType struct {
	ID      string
	Inward  string
	Outward string
}

// TrackerLinkedIssue represents a linked issue reference.
type TrackerLinkedIssue struct {
	Self    string
	ID      string
	Key     string
	Display string
}

// TrackerLink represents a link between two issues.
type TrackerLink struct {
	ID        string
	Self      string
	Type      *TrackerLinkType
	Direction string
	Object    *TrackerLinkedIssue
	CreatedBy *TrackerUser
	UpdatedBy *TrackerUser
	CreatedAt string
	UpdatedAt string
}

// TrackerChangelogFieldChange represents a single field change in changelog.
type TrackerChangelogFieldChange struct {
	Field string
	From  any
	To    any
}

// TrackerChangelogEntry represents a single changelog entry for an issue.
type TrackerChangelogEntry struct {
	ID        string
	Self      string
	Issue     *TrackerLinkedIssue
	UpdatedAt string
	UpdatedBy *TrackerUser
	Type      string
	Transport string
	Fields    []TrackerChangelogFieldChange
}

// TrackerProjectComment represents a comment on a project entity.
type TrackerProjectComment struct {
	ID        string
	LongID    string
	Self      string
	Text      string
	CreatedAt string
	UpdatedAt string
	CreatedBy *TrackerUser
	UpdatedBy *TrackerUser
}
