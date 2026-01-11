package tracker

import (
	"encoding/json"
	"testing"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueToTrackerIssue(t *testing.T) {
	t.Parallel()

	dto := issueDTO{
		Self:            "https://api.tracker.yandex.net/v2/issues/QUEUE-1",
		ID:              "123",
		Key:             "QUEUE-1",
		Version:         5,
		Summary:         "Test issue summary",
		Description:     "Test description",
		StatusStartTime: "2024-01-01T10:00:00.000+0000",
		CreatedAt:       "2024-01-01T09:00:00.000+0000",
		UpdatedAt:       "2024-01-02T10:00:00.000+0000",
		ResolvedAt:      "2024-01-03T10:00:00.000+0000",
		Status: &statusDTO{
			Self:    "https://api.tracker.yandex.net/v2/statuses/1",
			ID:      "1",
			Key:     "open",
			Display: "Open",
		},
		Type: &typeDTO{
			Self:    "https://api.tracker.yandex.net/v2/issuetypes/2",
			ID:      "2",
			Key:     "bug",
			Display: "Bug",
		},
		Priority: &prioDTO{
			Self:    "https://api.tracker.yandex.net/v2/priorities/3",
			ID:      "3",
			Key:     "normal",
			Display: "Normal",
		},
		Queue: &queueDTO{
			Self:           "https://api.tracker.yandex.net/v2/queues/QUEUE",
			ID:             "10",
			Key:            "QUEUE",
			Display:        "Test Queue",
			Name:           "",
			Version:        0,
			Lead:           nil,
			AssignAuto:     false,
			AllowExternals: false,
			DenyVoting:     false,
		},
		Assignee: &userDTO{
			Self:        "https://api.tracker.yandex.net/v2/users/111",
			ID:          "111",
			UID:         "",
			Login:       "",
			Display:     "John Doe",
			FirstName:   "",
			LastName:    "",
			Email:       "",
			CloudUID:    "",
			PassportUID: "",
		},
		CreatedBy: &userDTO{
			Self:        "https://api.tracker.yandex.net/v2/users/222",
			ID:          "222",
			UID:         "",
			Login:       "",
			Display:     "Jane Smith",
			FirstName:   "",
			LastName:    "",
			Email:       "",
			CloudUID:    "",
			PassportUID: "",
		},
		UpdatedBy: nil,
		Votes:     10,
		Favorite:  true,
	}

	result := issueToTrackerIssue(dto)

	assert.Equal(t, dto.Self, result.Self)
	assert.Equal(t, string(dto.ID), result.ID)
	assert.Equal(t, dto.Key, result.Key)
	assert.Equal(t, dto.Version, result.Version)
	assert.Equal(t, dto.Summary, result.Summary)
	assert.Equal(t, dto.Description, result.Description)
	assert.Equal(t, dto.StatusStartTime, result.StatusStartTime)
	assert.Equal(t, dto.CreatedAt, result.CreatedAt)
	assert.Equal(t, dto.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, dto.ResolvedAt, result.ResolvedAt)
	assert.Equal(t, dto.Votes, result.Votes)
	assert.Equal(t, dto.Favorite, result.Favorite)

	require.NotNil(t, result.Status)
	assert.Equal(t, dto.Status.Display, result.Status.Display)

	require.NotNil(t, result.Type)
	assert.Equal(t, dto.Type.Display, result.Type.Display)

	require.NotNil(t, result.Priority)
	assert.Equal(t, dto.Priority.Display, result.Priority.Display)

	require.NotNil(t, result.Queue)
	assert.Equal(t, dto.Queue.Key, result.Queue.Key)

	require.NotNil(t, result.Assignee)
	assert.Equal(t, dto.Assignee.Display, result.Assignee.Display)

	require.NotNil(t, result.CreatedBy)
	assert.Equal(t, dto.CreatedBy.Display, result.CreatedBy.Display)
}

func TestIssueToTrackerIssue_NilNestedObjects(t *testing.T) {
	t.Parallel()

	dto := issueDTO{
		Self:            "https://api.tracker.yandex.net/v2/issues/QUEUE-1",
		ID:              "123",
		Key:             "QUEUE-1",
		Version:         0,
		Summary:         "Minimal issue",
		Description:     "",
		StatusStartTime: "",
		CreatedAt:       "",
		UpdatedAt:       "",
		ResolvedAt:      "",
		Status:          nil,
		Type:            nil,
		Priority:        nil,
		Queue:           nil,
		Assignee:        nil,
		CreatedBy:       nil,
		UpdatedBy:       nil,
		Votes:           0,
		Favorite:        false,
	}

	result := issueToTrackerIssue(dto)

	assert.Equal(t, dto.Key, result.Key)
	assert.Nil(t, result.Status)
	assert.Nil(t, result.Type)
	assert.Nil(t, result.Priority)
	assert.Nil(t, result.Queue)
	assert.Nil(t, result.Assignee)
	assert.Nil(t, result.CreatedBy)
	assert.Nil(t, result.UpdatedBy)
}

func TestStatusToTrackerStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *statusDTO
		expected *domain.TrackerStatus
	}{
		{
			name:     "nil status",
			input:    nil,
			expected: nil,
		},
		{
			name: "full status",
			input: &statusDTO{
				Self:    "https://api.tracker.yandex.net/v2/statuses/1",
				ID:      "1",
				Key:     "open",
				Display: "Open",
			},
			expected: &domain.TrackerStatus{
				Self:    "https://api.tracker.yandex.net/v2/statuses/1",
				ID:      "1",
				Key:     "open",
				Display: "Open",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := statusToTrackerStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserToTrackerUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *userDTO
		expected *domain.TrackerUser
	}{
		{
			name:     "nil user",
			input:    nil,
			expected: nil,
		},
		{
			name: "full user",
			input: &userDTO{
				Self:        "https://api.tracker.yandex.net/v2/users/123",
				ID:          "123",
				UID:         "456",
				Login:       "john.doe",
				Display:     "John Doe",
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				CloudUID:    "cloud-789",
				PassportUID: "1001",
			},
			expected: &domain.TrackerUser{
				Self:        "https://api.tracker.yandex.net/v2/users/123",
				ID:          "123",
				UID:         "456",
				Login:       "john.doe",
				Display:     "John Doe",
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				CloudUID:    "cloud-789",
				PassportUID: "1001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := userToTrackerUser(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueueToTrackerQueue(t *testing.T) {
	t.Parallel()

	lead := &userDTO{
		Self:        "https://api.tracker.yandex.net/v2/users/100",
		ID:          "100",
		UID:         "",
		Login:       "",
		Display:     "Team Lead",
		FirstName:   "",
		LastName:    "",
		Email:       "",
		CloudUID:    "",
		PassportUID: "",
	}

	tests := []struct {
		name     string
		input    *queueDTO
		expected *domain.TrackerQueue
	}{
		{
			name:     "nil queue",
			input:    nil,
			expected: nil,
		},
		{
			name: "full queue with lead",
			input: &queueDTO{
				Self:           "https://api.tracker.yandex.net/v2/queues/PROJ",
				ID:             "50",
				Key:            "PROJ",
				Display:        "Project Queue",
				Name:           "Project Queue Name",
				Version:        3,
				Lead:           lead,
				AssignAuto:     true,
				AllowExternals: false,
				DenyVoting:     true,
			},
			expected: &domain.TrackerQueue{
				Self:    "https://api.tracker.yandex.net/v2/queues/PROJ",
				ID:      "50",
				Key:     "PROJ",
				Display: "Project Queue",
				Name:    "Project Queue Name",
				Version: 3,
				Lead: &domain.TrackerUser{
					Self:        lead.Self,
					ID:          lead.ID.String(),
					UID:         "",
					Login:       "",
					Display:     lead.Display,
					FirstName:   "",
					LastName:    "",
					Email:       "",
					CloudUID:    "",
					PassportUID: "",
				},
				AssignAuto:     true,
				AllowExternals: false,
				DenyVoting:     true,
			},
		},
		{
			name: "queue without lead",
			input: &queueDTO{
				Self:           "https://api.tracker.yandex.net/v2/queues/TEST",
				ID:             "60",
				Key:            "TEST",
				Display:        "",
				Name:           "",
				Version:        0,
				Lead:           nil,
				AssignAuto:     false,
				AllowExternals: false,
				DenyVoting:     false,
			},
			expected: &domain.TrackerQueue{
				Self:           "https://api.tracker.yandex.net/v2/queues/TEST",
				ID:             "60",
				Key:            "TEST",
				Display:        "",
				Name:           "",
				Version:        0,
				Lead:           nil,
				AssignAuto:     false,
				AllowExternals: false,
				DenyVoting:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := queueToTrackerQueue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentToTrackerComment(t *testing.T) {
	t.Parallel()

	dto := commentDTO{
		ID:        "999",
		LongID:    "long-id-999",
		Self:      "https://api.tracker.yandex.net/v2/issues/QUEUE-1/comments/999",
		Text:      "This is a test comment",
		Version:   2,
		Type:      "standard",
		Transport: "internal",
		CreatedAt: "2024-01-05T11:00:00.000+0000",
		UpdatedAt: "2024-01-05T12:00:00.000+0000",
		CreatedBy: &userDTO{
			Self:        "https://api.tracker.yandex.net/v2/users/300",
			ID:          "300",
			UID:         "",
			Login:       "",
			Display:     "Commenter",
			FirstName:   "",
			LastName:    "",
			Email:       "",
			CloudUID:    "",
			PassportUID: "",
		},
		UpdatedBy: &userDTO{
			Self:        "https://api.tracker.yandex.net/v2/users/301",
			ID:          "301",
			UID:         "",
			Login:       "",
			Display:     "Editor",
			FirstName:   "",
			LastName:    "",
			Email:       "",
			CloudUID:    "",
			PassportUID: "",
		},
	}

	result := commentToTrackerComment(dto)

	assert.Equal(t, string(dto.ID), result.ID)
	assert.EqualValues(t, dto.LongID, result.LongID)
	assert.Equal(t, dto.Self, result.Self)
	assert.Equal(t, dto.Text, result.Text)
	assert.Equal(t, dto.Version, result.Version)
	assert.Equal(t, dto.Type, result.Type)
	assert.Equal(t, dto.Transport, result.Transport)
	assert.Equal(t, dto.CreatedAt, result.CreatedAt)
	assert.Equal(t, dto.UpdatedAt, result.UpdatedAt)

	require.NotNil(t, result.CreatedBy)
	assert.Equal(t, dto.CreatedBy.Display, result.CreatedBy.Display)

	require.NotNil(t, result.UpdatedBy)
	assert.Equal(t, dto.UpdatedBy.Display, result.UpdatedBy.Display)
}

func TestTransitionToTrackerTransition(t *testing.T) {
	t.Parallel()

	dto := transitionDTO{
		ID:      "tr-1",
		Display: "Close",
		Self:    "https://api.tracker.yandex.net/v2/issues/QUEUE-1/transitions/tr-1",
		To: &statusDTO{
			Self:    "https://api.tracker.yandex.net/v2/statuses/closed",
			ID:      "closed-id",
			Key:     "closed",
			Display: "Closed",
		},
	}

	result := transitionToTrackerTransition(dto)

	assert.Equal(t, string(dto.ID), result.ID)
	assert.Equal(t, dto.Display, result.Display)
	assert.Equal(t, dto.Self, result.Self)
	require.NotNil(t, result.To)
	assert.Equal(t, dto.To.Key, result.To.Key)
	assert.Equal(t, dto.To.Display, result.To.Display)
}

func TestSearchIssuesResultToTrackerIssuesPage(t *testing.T) {
	t.Parallel()

	issues := []issueDTO{
		{
			Self:            "https://api.tracker.yandex.net/v2/issues/QUEUE-1",
			ID:              "1",
			Key:             "QUEUE-1",
			Version:         0,
			Summary:         "First issue",
			Description:     "",
			StatusStartTime: "",
			CreatedAt:       "",
			UpdatedAt:       "",
			ResolvedAt:      "",
			Status: &statusDTO{
				Self:    "https://api.tracker.yandex.net/v2/statuses/open",
				ID:      "1",
				Key:     "open",
				Display: "Open",
			},
			Type:      nil,
			Priority:  nil,
			Queue:     nil,
			Assignee:  nil,
			CreatedBy: nil,
			UpdatedBy: nil,
			Votes:     0,
			Favorite:  false,
		},
		{
			Self:            "https://api.tracker.yandex.net/v2/issues/QUEUE-2",
			ID:              "2",
			Key:             "QUEUE-2",
			Version:         0,
			Summary:         "Second issue",
			Description:     "",
			StatusStartTime: "",
			CreatedAt:       "",
			UpdatedAt:       "",
			ResolvedAt:      "",
			Status:          nil,
			Type:            nil,
			Priority:        nil,
			Queue:           nil,
			Assignee:        nil,
			CreatedBy:       nil,
			UpdatedBy:       nil,
			Votes:           0,
			Favorite:        false,
		},
	}
	totalCount := 100
	totalPages := 10
	scrollID := "scroll-abc"
	scrollToken := "token-xyz"
	nextLink := "https://api.tracker.yandex.net/v2/issues?page=2"

	result := searchIssuesResultToTrackerIssuesPage(issues, totalCount, totalPages, scrollID, scrollToken, nextLink)

	assert.Len(t, result.Issues, 2)
	assert.Equal(t, issues[0].Key, result.Issues[0].Key)
	assert.Equal(t, issues[0].Summary, result.Issues[0].Summary)
	require.NotNil(t, result.Issues[0].Status)
	assert.Equal(t, issues[0].Status.Display, result.Issues[0].Status.Display)

	assert.Equal(t, issues[1].Key, result.Issues[1].Key)
	assert.Nil(t, result.Issues[1].Status)

	assert.Equal(t, totalCount, result.TotalCount)
	assert.Equal(t, totalPages, result.TotalPages)
	assert.Equal(t, scrollID, result.ScrollID)
	assert.Equal(t, scrollToken, result.ScrollToken)
	assert.Equal(t, nextLink, result.NextLink)
}

func TestSearchIssuesResultToTrackerIssuesPage_EmptyIssues(t *testing.T) {
	t.Parallel()

	issues := []issueDTO{}
	totalCount := 0
	totalPages := 0
	scrollID := ""
	scrollToken := ""
	nextLink := ""

	result := searchIssuesResultToTrackerIssuesPage(issues, totalCount, totalPages, scrollID, scrollToken, nextLink)

	assert.Empty(t, result.Issues)
	assert.Zero(t, result.TotalCount)
	assert.Zero(t, result.TotalPages)
}

func TestListQueuesResultToTrackerQueuesPage(t *testing.T) {
	t.Parallel()

	lead := &userDTO{
		Self:        "https://api.tracker.yandex.net/v2/users/lead",
		ID:          "lead-id",
		UID:         "",
		Login:       "",
		Display:     "Queue Lead",
		FirstName:   "",
		LastName:    "",
		Email:       "",
		CloudUID:    "",
		PassportUID: "",
	}

	queues := []queueDTO{
		{
			Self:           "https://api.tracker.yandex.net/v2/queues/PROJ1",
			ID:             "1",
			Key:            "PROJ1",
			Display:        "Project One",
			Name:           "Project One Name",
			Version:        5,
			Lead:           lead,
			AssignAuto:     false,
			AllowExternals: false,
			DenyVoting:     false,
		},
		{
			Self:           "https://api.tracker.yandex.net/v2/queues/PROJ2",
			ID:             "2",
			Key:            "PROJ2",
			Display:        "Project Two",
			Name:           "",
			Version:        0,
			Lead:           nil,
			AssignAuto:     false,
			AllowExternals: false,
			DenyVoting:     false,
		},
	}
	totalCount := 50
	totalPages := 5

	result := listQueuesResultToTrackerQueuesPage(queues, totalCount, totalPages)

	assert.Len(t, result.Queues, 2)
	assert.Equal(t, queues[0].Key, result.Queues[0].Key)
	assert.Equal(t, queues[0].Name, result.Queues[0].Name)
	require.NotNil(t, result.Queues[0].Lead)
	assert.Equal(t, lead.Display, result.Queues[0].Lead.Display)

	assert.Equal(t, queues[1].Key, result.Queues[1].Key)
	assert.Nil(t, result.Queues[1].Lead)

	assert.Equal(t, totalCount, result.TotalCount)
	assert.Equal(t, totalPages, result.TotalPages)
}

func TestQueueUnmarshalJSONWithNumericID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonData string
		wantID   string
		wantErr  bool
	}{
		{
			name:     "numeric ID as float64",
			jsonData: `{"self":"https://api.tracker.yandex.net/v2/queues/PROJ","id":123,"key":"PROJ"}`,
			wantID:   "123",
			wantErr:  false,
		},
		{
			name:     "string ID",
			jsonData: `{"self":"https://api.tracker.yandex.net/v2/queues/PROJ","id":"my-string-id","key":"PROJ"}`,
			wantID:   "my-string-id",
			wantErr:  false,
		},
		{
			name:     "numeric string ID",
			jsonData: `{"self":"https://api.tracker.yandex.net/v2/queues/PROJ","id":"789","key":"PROJ"}`,
			wantID:   "789",
			wantErr:  false,
		},
		{
			name:     "unsupported type (boolean)",
			jsonData: `{"self":"https://api.tracker.yandex.net/v2/queues/PROJ","id":true,"key":"PROJ"}`,
			wantID:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var q queueDTO
			err := json.Unmarshal([]byte(tt.jsonData), &q)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported queue ID type")
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tt.wantID, q.ID)
			}
		})
	}
}

func TestListCommentsResultToTrackerCommentsPage(t *testing.T) {
	t.Parallel()

	comments := []commentDTO{
		{
			ID:        "1",
			LongID:    "long-1",
			Self:      "https://api.tracker.yandex.net/v2/issues/QUEUE-1/comments/1",
			Text:      "First comment",
			Version:   0,
			Type:      "",
			Transport: "",
			CreatedAt: "",
			UpdatedAt: "",
			CreatedBy: &userDTO{
				Self:        "https://api.tracker.yandex.net/v2/users/author",
				ID:          "author-id",
				UID:         "",
				Login:       "",
				Display:     "Author",
				FirstName:   "",
				LastName:    "",
				Email:       "",
				CloudUID:    "",
				PassportUID: "",
			},
			UpdatedBy: nil,
		},
		{
			ID:        "2",
			LongID:    "long-2",
			Self:      "https://api.tracker.yandex.net/v2/issues/QUEUE-1/comments/2",
			Text:      "Second comment",
			Version:   0,
			Type:      "",
			Transport: "",
			CreatedAt: "",
			UpdatedAt: "",
			CreatedBy: nil,
			UpdatedBy: nil,
		},
	}
	nextLink := "https://api.tracker.yandex.net/v2/issues/QUEUE-1/comments?id=3"

	result := listCommentsResultToTrackerCommentsPage(comments, nextLink)

	assert.Len(t, result.Comments, 2)
	assert.EqualValues(t, comments[0].ID, result.Comments[0].ID)
	assert.Equal(t, comments[0].Text, result.Comments[0].Text)
	require.NotNil(t, result.Comments[0].CreatedBy)
	assert.Equal(t, comments[0].CreatedBy.Display, result.Comments[0].CreatedBy.Display)

	assert.EqualValues(t, comments[1].ID, result.Comments[1].ID)
	assert.Nil(t, result.Comments[1].CreatedBy)

	assert.Equal(t, nextLink, result.NextLink)
}

func TestPrioToTrackerPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *prioDTO
		expected *domain.TrackerPriority
	}{
		{
			name:     "nil priority",
			input:    nil,
			expected: nil,
		},
		{
			name: "full priority",
			input: &prioDTO{
				Self:    "https://api.tracker.yandex.net/v2/priorities/critical",
				ID:      "critical-id",
				Key:     "critical",
				Display: "Critical",
			},
			expected: &domain.TrackerPriority{
				Self:    "https://api.tracker.yandex.net/v2/priorities/critical",
				ID:      "critical-id",
				Key:     "critical",
				Display: "Critical",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := prioToTrackerPriority(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeToTrackerIssueType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *typeDTO
		expected *domain.TrackerIssueType
	}{
		{
			name:     "nil type",
			input:    nil,
			expected: nil,
		},
		{
			name: "full type",
			input: &typeDTO{
				Self:    "https://api.tracker.yandex.net/v2/issuetypes/task",
				ID:      "task-id",
				Key:     "task",
				Display: "Task",
			},
			expected: &domain.TrackerIssueType{
				Self:    "https://api.tracker.yandex.net/v2/issuetypes/task",
				ID:      "task-id",
				Key:     "task",
				Display: "Task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := typeToTrackerIssueType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
