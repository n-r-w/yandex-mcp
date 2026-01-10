package tracker

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
)

// issueDTO is a Yandex Tracker issue.
type issueDTO struct {
	Self            string              `json:"self"`
	ID              apihelpers.StringID `json:"id"`
	Key             string              `json:"key"`
	Version         int                 `json:"version"`
	Summary         string              `json:"summary"`
	Description     string              `json:"description,omitempty"`
	StatusStartTime string              `json:"statusStartTime,omitempty"`
	CreatedAt       string              `json:"createdAt,omitempty"`
	UpdatedAt       string              `json:"updatedAt,omitempty"`
	ResolvedAt      string              `json:"resolvedAt,omitempty"`
	Status          *statusDTO          `json:"status,omitempty"`
	Type            *typeDTO            `json:"type,omitempty"`
	Priority        *prioDTO            `json:"priority,omitempty"`
	Queue           *queueDTO           `json:"queue,omitempty"`
	Assignee        *userDTO            `json:"assignee,omitempty"`
	CreatedBy       *userDTO            `json:"createdBy,omitempty"`
	UpdatedBy       *userDTO            `json:"updatedBy,omitempty"`
	Votes           int                 `json:"votes,omitempty"`
	Favorite        bool                `json:"favorite,omitempty"`
}

// statusDTO represents an issue status.
type statusDTO struct {
	Self    string              `json:"self"`
	ID      apihelpers.StringID `json:"id"`
	Key     string              `json:"key"`
	Display string              `json:"display"`
}

// typeDTO represents an issue type.
type typeDTO struct {
	Self    string              `json:"self"`
	ID      apihelpers.StringID `json:"id"`
	Key     string              `json:"key"`
	Display string              `json:"display"`
}

// prioDTO represents an issue priority.
type prioDTO struct {
	Self    string              `json:"self"`
	ID      apihelpers.StringID `json:"id"`
	Key     string              `json:"key"`
	Display string              `json:"display"`
}

// queueDTO represents a Tracker queue.
type queueDTO struct {
	Self           string              `json:"self"`
	ID             apihelpers.StringID `json:"id"`
	Key            string              `json:"key"`
	Display        string              `json:"display,omitempty"`
	Name           string              `json:"name,omitempty"`
	Version        int                 `json:"version,omitempty"`
	Lead           *userDTO            `json:"lead,omitempty"`
	AssignAuto     bool                `json:"assignAuto,omitempty"`
	AllowExternals bool                `json:"allowExternals,omitempty"`
	DenyVoting     bool                `json:"denyVoting,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Queue to handle numeric and string IDs.
func (q *queueDTO) UnmarshalJSON(data []byte) error {
	type QueueAlias queueDTO

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
			q.ID = apihelpers.StringID(strconv.FormatFloat(v, 'f', -1, 64))
		case string:
			q.ID = apihelpers.StringID(v)
		default:
			return fmt.Errorf("unsupported queue ID type: %T", v)
		}
	}

	return nil
}

// userDTO represents a Tracker user.
type userDTO struct {
	Self        string              `json:"self"`
	ID          apihelpers.StringID `json:"id"`
	UID         apihelpers.StringID `json:"uid,omitempty"`
	Login       string              `json:"login,omitempty"`
	Display     string              `json:"display,omitempty"`
	FirstName   string              `json:"firstName,omitempty"`
	LastName    string              `json:"lastName,omitempty"`
	Email       string              `json:"email,omitempty"`
	CloudUID    apihelpers.StringID `json:"cloudUid,omitempty"`
	PassportUID apihelpers.StringID `json:"passportUid,omitempty"`
}

// transitionDTO represents an available issue transition.
type transitionDTO struct {
	ID      apihelpers.StringID `json:"id"`
	Display string              `json:"display"`
	Self    string              `json:"self"`
	To      *statusDTO          `json:"to"`
}

// commentDTO represents an issue comment.
type commentDTO struct {
	ID        apihelpers.StringID `json:"id"`
	LongID    apihelpers.StringID `json:"longId"`
	Self      string              `json:"self"`
	Text      string              `json:"text"`
	Version   int                 `json:"version"`
	Type      string              `json:"type,omitempty"`
	Transport string              `json:"transport,omitempty"`
	CreatedAt string              `json:"createdAt,omitempty"`
	UpdatedAt string              `json:"updatedAt,omitempty"`
	CreatedBy *userDTO            `json:"createdBy,omitempty"`
	UpdatedBy *userDTO            `json:"updatedBy,omitempty"`
}

// searchRequestDTO represents the request body for issue search.
type searchRequestDTO struct {
	Filter map[string]any `json:"filter,omitempty"`
	Query  string         `json:"query,omitempty"`
	Order  string         `json:"order,omitempty"`
}

// countRequestDTO represents the request body for issue count.
type countRequestDTO struct {
	Filter map[string]any `json:"filter,omitempty"`
	Query  string         `json:"query,omitempty"`
}

// errorResponseDTO represents the Tracker API error format.
type errorResponseDTO struct {
	Errors        []string `json:"errors,omitempty"`
	ErrorMessages []string `json:"errorMessages,omitempty"`
	StatusCode    int      `json:"statusCode,omitempty"`
}

// Write operation request DTOs.

// createIssueRequestDTO is the request body for issue creation.
type createIssueRequestDTO struct {
	Queue         string                `json:"queue"`
	Summary       string                `json:"summary"`
	Description   string                `json:"description,omitempty"`
	Type          string                `json:"type,omitempty"`
	Priority      string                `json:"priority,omitempty"`
	Assignee      string                `json:"assignee,omitempty"`
	Tags          []string              `json:"tags,omitempty"`
	Parent        string                `json:"parent,omitempty"`
	AttachmentIDs []apihelpers.StringID `json:"attachmentIds,omitempty"`
	Sprint        []string              `json:"sprint,omitempty"`
}

// updateIssueProjectDTO is the nested project structure for issue update.
type updateIssueProjectDTO struct {
	Primary   int                          `json:"primary,omitempty"`
	Secondary *updateIssueProjectSecAddDTO `json:"secondary,omitempty"`
}

// updateIssueProjectSecAddDTO is the secondary project add structure.
type updateIssueProjectSecAddDTO struct {
	Add []int `json:"add,omitempty"`
}

// updateIssueSprintDTO is the sprint reference for issue update.
type updateIssueSprintDTO struct {
	ID apihelpers.StringID `json:"id"`
}

// updateIssueRequestDTO is the request body for issue update.
type updateIssueRequestDTO struct {
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	Assignee    string                 `json:"assignee,omitempty"`
	Version     int                    `json:"version,omitempty"`
	Project     *updateIssueProjectDTO `json:"project,omitempty"`
	Sprint      []updateIssueSprintDTO `json:"sprint,omitempty"`
}

// executeTransitionRequestDTO is the request body for transition execution.
type executeTransitionRequestDTO struct {
	Comment string         `json:"comment,omitempty"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// addCommentRequestDTO is the request body for adding a comment.
type addCommentRequestDTO struct {
	Text              string                `json:"text"`
	AttachmentIDs     []apihelpers.StringID `json:"attachmentIds,omitempty"`
	MarkupType        string                `json:"markupType,omitempty"`
	Summonees         []string              `json:"summonees,omitempty"`
	MaillistSummonees []string              `json:"maillistSummonees,omitempty"`
	IsAddToFollowers  *bool                 `json:"isAddToFollowers,omitempty"`
}

// updateCommentRequestDTO is the request body for updating a comment.
type updateCommentRequestDTO struct {
	Text              string                `json:"text"`
	AttachmentIDs     []apihelpers.StringID `json:"attachmentIds,omitempty"`
	MarkupType        string                `json:"markupType,omitempty"`
	Summonees         []string              `json:"summonees,omitempty"`
	MaillistSummonees []string              `json:"maillistSummonees,omitempty"`
}

// attachmentDTO represents a file attachment in the Tracker API.
type attachmentDTO struct {
	ID        apihelpers.StringID    `json:"id"`
	Name      string                 `json:"name"`
	Content   string                 `json:"content"`
	Thumbnail string                 `json:"thumbnail,omitempty"`
	Mimetype  string                 `json:"mimetype,omitempty"`
	Size      int64                  `json:"size"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	CreatedBy *userDTO               `json:"createdBy,omitempty"`
	Metadata  *attachmentMetadataDTO `json:"metadata,omitempty"`
}

// attachmentMetadataDTO contains additional attachment metadata.
type attachmentMetadataDTO struct {
	Size string `json:"size,omitempty"`
}

// queueDetailDTO represents a detailed queue response from the Tracker API.
type queueDetailDTO struct {
	Self            string              `json:"self"`
	ID              apihelpers.StringID `json:"id"`
	Key             string              `json:"key"`
	Display         string              `json:"display,omitempty"`
	Name            string              `json:"name,omitempty"`
	Description     string              `json:"description,omitempty"`
	Version         int                 `json:"version,omitempty"`
	Lead            *userDTO            `json:"lead,omitempty"`
	AssignAuto      bool                `json:"assignAuto,omitempty"`
	AllowExternals  bool                `json:"allowExternals,omitempty"`
	DenyVoting      bool                `json:"denyVoting,omitempty"`
	DefaultType     *typeDTO            `json:"defaultType,omitempty"`
	DefaultPriority *prioDTO            `json:"defaultPriority,omitempty"`
}

// userDetailDTO represents a detailed user response from the Tracker API.
type userDetailDTO struct {
	Self        string              `json:"self"`
	ID          apihelpers.StringID `json:"id"`
	UID         apihelpers.StringID `json:"uid,omitempty"`
	TrackerUID  apihelpers.StringID `json:"trackerUid,omitempty"`
	Login       string              `json:"login,omitempty"`
	Display     string              `json:"display,omitempty"`
	FirstName   string              `json:"firstName,omitempty"`
	LastName    string              `json:"lastName,omitempty"`
	Email       string              `json:"email,omitempty"`
	CloudUID    apihelpers.StringID `json:"cloudUid,omitempty"`
	PassportUID apihelpers.StringID `json:"passportUid,omitempty"`
	HasLicense  bool                `json:"hasLicense,omitempty"`
	Dismissed   bool                `json:"dismissed,omitempty"`
	External    bool                `json:"external,omitempty"`
}

// createQueueRequestDTO is the request body for queue creation.
type createQueueRequestDTO struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Lead            string `json:"lead"`
	DefaultType     string `json:"defaultType"`
	DefaultPriority string `json:"defaultPriority"`
}

// linkDTO represents an issue link in the Tracker API.
type linkDTO struct {
	ID        apihelpers.StringID `json:"id"`
	Self      string              `json:"self"`
	Type      *linkTypeDTO        `json:"type,omitempty"`
	Direction string              `json:"direction,omitempty"`
	Object    *linkedIssueDTO     `json:"object,omitempty"`
	CreatedBy *userDTO            `json:"createdBy,omitempty"`
	UpdatedBy *userDTO            `json:"updatedBy,omitempty"`
	CreatedAt string              `json:"createdAt,omitempty"`
	UpdatedAt string              `json:"updatedAt,omitempty"`
}

// linkTypeDTO represents a link type in the Tracker API.
type linkTypeDTO struct {
	ID      apihelpers.StringID `json:"id"`
	Inward  string              `json:"inward,omitempty"`
	Outward string              `json:"outward,omitempty"`
}

// linkedIssueDTO represents a linked issue reference in the Tracker API.
type linkedIssueDTO struct {
	Self    string              `json:"self"`
	ID      apihelpers.StringID `json:"id"`
	Key     string              `json:"key"`
	Display string              `json:"display,omitempty"`
}

// createLinkRequestDTO is the request body for creating a link.
type createLinkRequestDTO struct {
	Relationship string `json:"relationship"`
	Issue        string `json:"issue"`
}

// changelogEntryDTO represents a single changelog entry in the Tracker API.
type changelogEntryDTO struct {
	ID        apihelpers.StringID       `json:"id"`
	Self      string                    `json:"self"`
	Issue     *linkedIssueDTO           `json:"issue,omitempty"`
	UpdatedAt string                    `json:"updatedAt,omitempty"`
	UpdatedBy *userDTO                  `json:"updatedBy,omitempty"`
	Type      string                    `json:"type,omitempty"`
	Transport string                    `json:"transport,omitempty"`
	Fields    []changelogFieldChangeDTO `json:"fields,omitempty"`
}

// changelogFieldDTO represents the field object in changelog entries.
type changelogFieldDTO struct {
	Self    string              `json:"self"`
	ID      apihelpers.StringID `json:"id"`
	Display string              `json:"display"`
}

// changelogFieldChangeDTO represents a single field change in a changelog entry.
type changelogFieldChangeDTO struct {
	Field *changelogFieldDTO `json:"field"`
	From  any                `json:"from,omitempty"`
	To    any                `json:"to,omitempty"`
}

// projectCommentDTO represents a project entity comment in the Tracker API.
type projectCommentDTO struct {
	ID        apihelpers.StringID `json:"id"`
	LongID    apihelpers.StringID `json:"longId,omitempty"`
	Self      string              `json:"self"`
	Text      string              `json:"text,omitempty"`
	CreatedAt string              `json:"createdAt,omitempty"`
	UpdatedAt string              `json:"updatedAt,omitempty"`
	CreatedBy *userDTO            `json:"createdBy,omitempty"`
	UpdatedBy *userDTO            `json:"updatedBy,omitempty"`
}
