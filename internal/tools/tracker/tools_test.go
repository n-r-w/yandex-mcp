//nolint:exhaustruct // test file uses partial struct initialization for clarity
package tracker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

func TestTools_GetIssue(t *testing.T) {
	t.Parallel()

	t.Run("returns error when issue_id_or_key is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getIssue(context.Background(), getIssueInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedIssue := &domain.TrackerIssue{
			Self:    "https://api.tracker/v3/issues/TEST-123",
			ID:      "12345",
			Key:     "TEST-123",
			Version: 1,
			Summary: "Test Issue",
			Status:  &domain.TrackerStatus{ID: "1", Key: "open", Display: "Open"},
		}

		mockAdapter.EXPECT().
			GetIssue(gomock.Any(), "TEST-123", domain.TrackerGetIssueOpts{Expand: "attachments"}).
			Return(expectedIssue, nil)

		result, err := reg.getIssue(context.Background(), getIssueInputDTO{
			IssueID: "TEST-123",
			Expand:  "attachments",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-123", result.Key)
		assert.Equal(t, "Test Issue", result.Summary)
		require.NotNil(t, result.Status)
		assert.Equal(t, "Open", result.Status.Display)
	})

	t.Run("returns safe error on upstream error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.UpstreamError{
			Service:    domain.ServiceTracker,
			Operation:  "GetIssue",
			HTTPStatus: 404,
			Message:    "Issue not found",
		}

		mockAdapter.EXPECT().
			GetIssue(gomock.Any(), "MISSING-1", domain.TrackerGetIssueOpts{}).
			Return(nil, upstreamErr)

		_, err := reg.getIssue(context.Background(), getIssueInputDTO{IssueID: "MISSING-1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), domain.ServiceTracker)
		assert.Contains(t, err.Error(), "GetIssue")
		assert.Contains(t, err.Error(), "HTTP 404")
	})
}

func TestTools_SearchIssues(t *testing.T) {
	t.Parallel()

	t.Run("returns error when per_page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.searchIssues(context.Background(), searchIssuesInputDTO{PerPage: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("returns error when page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.searchIssues(context.Background(), searchIssuesInputDTO{Page: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be non-negative")
	})

	t.Run("returns error when per_scroll is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.searchIssues(context.Background(), searchIssuesInputDTO{PerScroll: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_scroll must be non-negative")
	})

	t.Run("returns error when per_scroll exceeds max", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.searchIssues(context.Background(), searchIssuesInputDTO{PerScroll: 1001})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_scroll must not exceed 1000")
	})

	t.Run("returns error when scroll_ttl_millis is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.searchIssues(context.Background(), searchIssuesInputDTO{ScrollTTLMillis: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scroll_ttl_millis must be non-negative")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResult := &domain.TrackerIssuesPage{
			Issues: []domain.TrackerIssue{
				{ID: "1", Key: "TEST-1", Summary: "First"},
				{ID: "2", Key: "TEST-2", Summary: "Second"},
			},
			TotalCount:  100,
			TotalPages:  10,
			ScrollID:    "scroll123",
			ScrollToken: "token456",
			NextLink:    "https://api/next",
		}

		mockAdapter.EXPECT().
			SearchIssues(gomock.Any(), domain.TrackerSearchIssuesOpts{
				Filter:          map[string]string{"status": "open"},
				Query:           "Queue: TEST",
				Order:           "+updated",
				Expand:          "transitions",
				PerPage:         20,
				Page:            2,
				ScrollType:      "sorted",
				PerScroll:       100,
				ScrollTTLMillis: 5000,
				ScrollID:        "prevScroll",
			}).
			Return(expectedResult, nil)

		input := searchIssuesInputDTO{
			Filter:          map[string]string{"status": "open"},
			Query:           "Queue: TEST",
			Order:           "+updated",
			Expand:          "transitions",
			PerPage:         20,
			Page:            2,
			ScrollType:      "sorted",
			PerScroll:       100,
			ScrollTTLMillis: 5000,
			ScrollID:        "prevScroll",
		}

		result, err := reg.searchIssues(context.Background(), input)
		require.NoError(t, err)
		assert.Len(t, result.Issues, 2)
		assert.Equal(t, 100, result.TotalCount)
		assert.Equal(t, 10, result.TotalPages)
		assert.Equal(t, "scroll123", result.ScrollID)
		assert.Equal(t, "token456", result.ScrollToken)
		assert.Equal(t, "https://api/next", result.NextLink)
	})
}

func TestTools_CountIssues(t *testing.T) {
	t.Parallel()

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			CountIssues(gomock.Any(), domain.TrackerCountIssuesOpts{
				Filter: map[string]string{"assignee": "me"},
				Query:  "Queue: PROJ",
			}).
			Return(42, nil)

		result, err := reg.countIssues(context.Background(), countIssuesInputDTO{
			Filter: map[string]string{"assignee": "me"},
			Query:  "Queue: PROJ",
		})
		require.NoError(t, err)
		assert.Equal(t, 42, result.Count)
	})
}

func TestTools_ListTransitions(t *testing.T) {
	t.Parallel()

	t.Run("returns error when issue_id_or_key is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listTransitions(context.Background(), listTransitionsInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("calls adapter and maps result", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedTransitions := []domain.TrackerTransition{
			{
				ID:      "1",
				Display: "Start Work",
				Self:    "https://api/transitions/1",
				To:      &domain.TrackerStatus{ID: "2", Key: "inProgress", Display: "In Progress"},
			},
			{
				ID:      "2",
				Display: "Close",
				Self:    "https://api/transitions/2",
				To:      &domain.TrackerStatus{ID: "3", Key: "closed", Display: "Closed"},
			},
		}

		mockAdapter.EXPECT().
			ListIssueTransitions(gomock.Any(), "ISSUE-1").
			Return(expectedTransitions, nil)

		result, err := reg.listTransitions(context.Background(), listTransitionsInputDTO{
			IssueID: "ISSUE-1",
		})
		require.NoError(t, err)
		assert.Len(t, result.Transitions, 2)
		assert.Equal(t, "Start Work", result.Transitions[0].Display)
		require.NotNil(t, result.Transitions[0].To)
		assert.Equal(t, "In Progress", result.Transitions[0].To.Display)
	})
}

func TestTools_ListQueues(t *testing.T) {
	t.Parallel()

	t.Run("returns error when per_page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listQueues(context.Background(), listQueuesInputDTO{PerPage: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("returns error when page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listQueues(context.Background(), listQueuesInputDTO{Page: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be non-negative")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResult := &domain.TrackerQueuesPage{
			Queues: []domain.TrackerQueue{
				{ID: "1", Key: "PROJ1", Name: "Project 1"},
				{ID: "2", Key: "PROJ2", Name: "Project 2"},
			},
			TotalCount: 50,
			TotalPages: 5,
		}

		mockAdapter.EXPECT().
			ListQueues(gomock.Any(), domain.TrackerListQueuesOpts{
				Expand:  "lead",
				PerPage: 10,
				Page:    1,
			}).
			Return(expectedResult, nil)

		result, err := reg.listQueues(context.Background(), listQueuesInputDTO{
			Expand:  "lead",
			PerPage: 10,
			Page:    1,
		})
		require.NoError(t, err)
		assert.Len(t, result.Queues, 2)
		assert.Equal(t, "PROJ1", result.Queues[0].Key)
		assert.Equal(t, 50, result.TotalCount)
		assert.Equal(t, 5, result.TotalPages)
	})
}

func TestTools_ListComments(t *testing.T) {
	t.Parallel()

	t.Run("returns error when issue_id_or_key is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listComments(context.Background(), listCommentsInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("returns error when per_page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listComments(context.Background(), listCommentsInputDTO{
			IssueID: "TEST-1",
			PerPage: -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResult := &domain.TrackerCommentsPage{
			Comments: []domain.TrackerComment{
				{
					ID:        "1",
					LongID:    "longid1",
					Text:      "First comment",
					CreatedBy: &domain.TrackerUser{ID: "user1", Display: "User One"},
				},
				{
					ID:        "2",
					LongID:    "longid2",
					Text:      "Second comment",
					CreatedBy: &domain.TrackerUser{ID: "user2", Display: "User Two"},
				},
			},
			NextLink: "https://api/next",
		}

		mockAdapter.EXPECT().
			ListIssueComments(gomock.Any(), "TEST-1", domain.TrackerListCommentsOpts{
				Expand:  "html",
				PerPage: 20,
				ID:      "100",
			}).
			Return(expectedResult, nil)

		result, err := reg.listComments(context.Background(), listCommentsInputDTO{
			IssueID: "TEST-1",
			Expand:  "html",
			PerPage: 20,
			ID:      "100",
		})
		require.NoError(t, err)
		assert.Len(t, result.Comments, 2)
		assert.Equal(t, "First comment", result.Comments[0].Text)
		require.NotNil(t, result.Comments[0].CreatedBy)
		assert.Equal(t, "User One", result.Comments[0].CreatedBy.Display)
		assert.Equal(t, "https://api/next", result.NextLink)
	})
}

func TestTools_ErrorShaping(t *testing.T) {
	t.Parallel()

	t.Run("upstream error is shaped safely", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetIssue",
			500,
			"internal_error",
			"Internal server error",
			"body with Authorization: Bearer secret123",
		)

		mockAdapter.EXPECT().
			GetIssue(gomock.Any(), "TEST-1", domain.TrackerGetIssueOpts{}).
			Return(nil, upstreamErr)

		_, err := reg.getIssue(context.Background(), getIssueInputDTO{IssueID: "TEST-1"})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 500")
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret123")
	})

	t.Run("non-upstream error is shaped safely", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		// Simulate an error that contains sensitive data
		sensitiveErr := errors.New("connection failed: Authorization header: Bearer secret-token-123")

		mockAdapter.EXPECT().
			GetIssue(gomock.Any(), "TEST-1", domain.TrackerGetIssueOpts{}).
			Return(nil, sensitiveErr)

		_, err := reg.getIssue(context.Background(), getIssueInputDTO{IssueID: "TEST-1"})
		require.Error(t, err)
		errStr := err.Error()
		// Non-upstream errors should return a generic safe message
		assert.Equal(t, "tracker: internal error", errStr)
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret-token-123")
	})
}

func TestTools_MapsAllIssueFields(t *testing.T) {
	t.Parallel()

	t.Run("maps all issue fields correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedIssue := &domain.TrackerIssue{
			Self:            "https://api/issues/1",
			ID:              "1",
			Key:             "PROJ-1",
			Version:         5,
			Summary:         "Issue Summary",
			Description:     "Detailed description",
			StatusStartTime: "2024-01-01T10:00:00Z",
			CreatedAt:       "2024-01-01T00:00:00Z",
			UpdatedAt:       "2024-01-02T00:00:00Z",
			ResolvedAt:      "2024-01-03T00:00:00Z",
			Status:          &domain.TrackerStatus{ID: "s1", Key: "open", Display: "Open"},
			Type:            &domain.TrackerIssueType{ID: "t1", Key: "bug", Display: "Bug"},
			Priority:        &domain.TrackerPriority{ID: "p1", Key: "high", Display: "High"},
			Queue:           &domain.TrackerQueue{ID: "q1", Key: "PROJ", Name: "Project"},
			Assignee:        &domain.TrackerUser{ID: "u1", Display: "Assignee"},
			CreatedBy:       &domain.TrackerUser{ID: "u2", Display: "Creator"},
			UpdatedBy:       &domain.TrackerUser{ID: "u3", Display: "Updater"},
			Votes:           10,
			Favorite:        true,
		}

		mockAdapter.EXPECT().
			GetIssue(gomock.Any(), "PROJ-1", domain.TrackerGetIssueOpts{}).
			Return(expectedIssue, nil)

		result, err := reg.getIssue(context.Background(), getIssueInputDTO{IssueID: "PROJ-1"})
		require.NoError(t, err)

		assert.Equal(t, "https://api/issues/1", result.Self)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "PROJ-1", result.Key)
		assert.Equal(t, 5, result.Version)
		assert.Equal(t, "Issue Summary", result.Summary)
		assert.Equal(t, "Detailed description", result.Description)
		assert.Equal(t, "2024-01-01T10:00:00Z", result.StatusStartTime)
		assert.Equal(t, "2024-01-01T00:00:00Z", result.CreatedAt)
		assert.Equal(t, "2024-01-02T00:00:00Z", result.UpdatedAt)
		assert.Equal(t, "2024-01-03T00:00:00Z", result.ResolvedAt)

		require.NotNil(t, result.Status)
		assert.Equal(t, "Open", result.Status.Display)

		require.NotNil(t, result.Type)
		assert.Equal(t, "Bug", result.Type.Display)

		require.NotNil(t, result.Priority)
		assert.Equal(t, "High", result.Priority.Display)

		require.NotNil(t, result.Queue)
		assert.Equal(t, "Project", result.Queue.Name)

		require.NotNil(t, result.Assignee)
		assert.Equal(t, "Assignee", result.Assignee.Display)

		require.NotNil(t, result.CreatedBy)
		assert.Equal(t, "Creator", result.CreatedBy.Display)

		require.NotNil(t, result.UpdatedBy)
		assert.Equal(t, "Updater", result.UpdatedBy.Display)

		assert.Equal(t, 10, result.Votes)
		assert.True(t, result.Favorite)
	})
}

func TestTools_CreateIssue(t *testing.T) {
	t.Parallel()

	t.Run("validation/queue_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Summary: "Test Summary",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue is required")
	})

	t.Run("validation/summary_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Queue: "TEST",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "summary is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueCreateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/1",
				ID:      "1",
				Key:     "TEST-1",
				Summary: "Test Summary",
			},
		}

		mockAdapter.EXPECT().
			CreateIssue(gomock.Any(), &domain.TrackerIssueCreateRequest{
				Queue:   "TEST",
				Summary: "Test Summary",
			}).
			Return(expectedResp, nil)

		result, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Queue:   "TEST",
			Summary: "Test Summary",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
		assert.Equal(t, "Test Summary", result.Summary)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueCreateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/2",
				ID:      "2",
				Key:     "TEST-2",
				Summary: "Full Issue",
			},
		}

		mockAdapter.EXPECT().
			CreateIssue(gomock.Any(), &domain.TrackerIssueCreateRequest{
				Queue:         "TEST",
				Summary:       "Full Issue",
				Description:   "Issue description",
				Type:          "bug",
				Priority:      "critical",
				Assignee:      "user1",
				Tags:          []string{"tag1", "tag2"},
				Parent:        "TEST-1",
				AttachmentIDs: []string{"att1", "att2"},
				Sprint:        []string{"sprint1"},
			}).
			Return(expectedResp, nil)

		result, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Queue:         "TEST",
			Summary:       "Full Issue",
			Description:   "Issue description",
			Type:          "bug",
			Priority:      "critical",
			Assignee:      "user1",
			Tags:          []string{"tag1", "tag2"},
			Parent:        "TEST-1",
			AttachmentIDs: []string{"att1", "att2"},
			Sprint:        []string{"sprint1"},
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-2", result.Key)
	})

	t.Run("result/maps_issue_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueCreateResponse{
			Issue: domain.TrackerIssue{
				Self:      "https://api/issues/3",
				ID:        "3",
				Key:       "TEST-3",
				Version:   1,
				Summary:   "Created Issue",
				Status:    &domain.TrackerStatus{ID: "1", Key: "open", Display: "Open"},
				Type:      &domain.TrackerIssueType{ID: "t1", Key: "task", Display: "Task"},
				Priority:  &domain.TrackerPriority{ID: "p1", Key: "normal", Display: "Normal"},
				Queue:     &domain.TrackerQueue{ID: "q1", Key: "TEST", Name: "Test Queue"},
				CreatedBy: &domain.TrackerUser{ID: "u1", Display: "Creator"},
				CreatedAt: "2024-01-01T10:00:00Z",
			},
		}

		mockAdapter.EXPECT().
			CreateIssue(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Queue:   "TEST",
			Summary: "Created Issue",
		})
		require.NoError(t, err)
		assert.Equal(t, "https://api/issues/3", result.Self)
		assert.Equal(t, "3", result.ID)
		assert.Equal(t, "TEST-3", result.Key)
		assert.Equal(t, 1, result.Version)
		assert.Equal(t, "Created Issue", result.Summary)
		require.NotNil(t, result.Status)
		assert.Equal(t, "Open", result.Status.Display)
		require.NotNil(t, result.Type)
		assert.Equal(t, "Task", result.Type.Display)
		require.NotNil(t, result.Priority)
		assert.Equal(t, "Normal", result.Priority.Display)
		require.NotNil(t, result.Queue)
		assert.Equal(t, "Test Queue", result.Queue.Name)
		require.NotNil(t, result.CreatedBy)
		assert.Equal(t, "Creator", result.CreatedBy.Display)
		assert.Equal(t, "2024-01-01T10:00:00Z", result.CreatedAt)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"CreateIssue",
			400,
			"bad_request",
			"Queue not found",
			"body with Authorization: Bearer secret",
		)

		mockAdapter.EXPECT().
			CreateIssue(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.createIssue(context.Background(), createIssueInputDTO{
			Queue:   "INVALID",
			Summary: "Test",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 400")
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret")
	})
}

func TestTools_UpdateIssue(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			Summary: "Updated Summary",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/no_fields_to_update", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one field to update is required")
	})

	t.Run("adapter/call_with_summary_only", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/1",
				ID:      "1",
				Key:     "TEST-1",
				Summary: "Updated Summary",
			},
		}

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID: "TEST-1",
				Summary: "Updated Summary",
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "TEST-1",
			Summary: "Updated Summary",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Summary", result.Summary)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/1",
				ID:      "1",
				Key:     "TEST-1",
				Summary: "Full Update",
				Version: 2,
			},
		}

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID:     "TEST-1",
				Summary:     "Full Update",
				Description: "Updated description",
				Type:        "bug",
				Priority:    "high",
				Assignee:    "user2",
				Version:     1,
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID:     "TEST-1",
			Summary:     "Full Update",
			Description: "Updated description",
			Type:        "bug",
			Priority:    "high",
			Assignee:    "user2",
			Version:     1,
		})
		require.NoError(t, err)
		assert.Equal(t, "Full Update", result.Summary)
		assert.Equal(t, 2, result.Version)
	})

	t.Run("result/maps_issue_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{
				Self:      "https://api/issues/1",
				ID:        "1",
				Key:       "TEST-1",
				Version:   3,
				Summary:   "Updated",
				UpdatedAt: "2024-01-02T15:00:00Z",
				UpdatedBy: &domain.TrackerUser{ID: "u2", Display: "Updater"},
			},
		}

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "TEST-1",
			Summary: "Updated",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
		assert.Equal(t, 3, result.Version)
		assert.Equal(t, "2024-01-02T15:00:00Z", result.UpdatedAt)
		require.NotNil(t, result.UpdatedBy)
		assert.Equal(t, "Updater", result.UpdatedBy.Display)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"UpdateIssue",
			404,
			"not_found",
			"Issue not found",
			"body with token",
		)

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "MISSING-1",
			Summary: "Updated",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
	})

	t.Run("update_issue/project_primary_alone_is_valid", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{Key: "TEST-1"},
		}
		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID:        "TEST-1",
				ProjectPrimary: 123,
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID:        "TEST-1",
			ProjectPrimary: 123,
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
	})

	t.Run("update_issue/project_secondary_add_alone_is_valid", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{Key: "TEST-1"},
		}
		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID:             "TEST-1",
				ProjectSecondaryAdd: []int{456, 789},
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID:             "TEST-1",
			ProjectSecondaryAdd: []int{456, 789},
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
	})

	t.Run("update_issue/sprint_alone_is_valid", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{Key: "TEST-1"},
		}
		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID: "TEST-1",
				Sprint:  []string{"sprint-1", "sprint-2"},
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "TEST-1",
			Sprint:  []string{"sprint-1", "sprint-2"},
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
	})

	t.Run("update_issue/all_new_fields_together", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/1",
				ID:      "1",
				Key:     "TEST-1",
				Summary: "Issue with all new fields",
			},
		}

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID:             "TEST-1",
				ProjectPrimary:      123,
				ProjectSecondaryAdd: []int{456, 789},
				Sprint:              []string{"sprint-1"},
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID:             "TEST-1",
			ProjectPrimary:      123,
			ProjectSecondaryAdd: []int{456, 789},
			Sprint:              []string{"sprint-1"},
		})

		require.NoError(t, err)
		assert.Equal(t, "TEST-1", result.Key)
	})

	t.Run("update_issue/summary_only", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueUpdateResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/issues/1",
				ID:      "1",
				Key:     "TEST-1",
				Summary: "Updated Summary",
			},
		}

		mockAdapter.EXPECT().
			UpdateIssue(gomock.Any(), &domain.TrackerIssueUpdateRequest{
				IssueID: "TEST-1",
				Summary: "Updated Summary",
			}).
			Return(expectedResp, nil)

		result, err := reg.updateIssue(context.Background(), updateIssueInputDTO{
			IssueID: "TEST-1",
			Summary: "Updated Summary",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Summary", result.Summary)
	})
}

func TestTools_ExecuteTransition(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			TransitionID: "transition-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/transition_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transition_id is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerTransitionExecuteResponse{
			Transitions: []domain.TrackerTransition{
				{
					ID:      "2",
					Display: "Close",
					Self:    "https://api/transitions/2",
					To:      &domain.TrackerStatus{ID: "3", Key: "closed", Display: "Closed"},
				},
			},
		}

		mockAdapter.EXPECT().
			ExecuteTransition(gomock.Any(), &domain.TrackerTransitionExecuteRequest{
				IssueID:      "TEST-1",
				TransitionID: "transition-1",
			}).
			Return(expectedResp, nil)

		result, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			IssueID:      "TEST-1",
			TransitionID: "transition-1",
		})
		require.NoError(t, err)
		assert.Len(t, result.Transitions, 1)
		assert.Equal(t, "Close", result.Transitions[0].Display)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerTransitionExecuteResponse{
			Transitions: []domain.TrackerTransition{
				{
					ID:      "3",
					Display: "Reopen",
					Self:    "https://api/transitions/3",
				},
			},
		}

		mockAdapter.EXPECT().
			ExecuteTransition(gomock.Any(), &domain.TrackerTransitionExecuteRequest{
				IssueID:      "TEST-1",
				TransitionID: "transition-2",
				Comment:      "Transition comment",
				Fields:       map[string]any{"priority": "critical"},
			}).
			Return(expectedResp, nil)

		result, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			IssueID:      "TEST-1",
			TransitionID: "transition-2",
			Comment:      "Transition comment",
			Fields:       map[string]any{"priority": "critical"},
		})
		require.NoError(t, err)
		assert.Len(t, result.Transitions, 1)
	})

	t.Run("result/maps_transition_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerTransitionExecuteResponse{
			Transitions: []domain.TrackerTransition{
				{
					ID:      "1",
					Display: "Start Progress",
					Self:    "https://api/transitions/1",
					To:      &domain.TrackerStatus{ID: "2", Key: "inProgress", Display: "In Progress"},
				},
				{
					ID:      "2",
					Display: "Close",
					Self:    "https://api/transitions/2",
					To:      &domain.TrackerStatus{ID: "3", Key: "closed", Display: "Closed"},
				},
			},
		}

		mockAdapter.EXPECT().
			ExecuteTransition(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			IssueID:      "TEST-1",
			TransitionID: "transition-1",
		})
		require.NoError(t, err)
		require.Len(t, result.Transitions, 2)
		assert.Equal(t, "1", result.Transitions[0].ID)
		assert.Equal(t, "Start Progress", result.Transitions[0].Display)
		require.NotNil(t, result.Transitions[0].To)
		assert.Equal(t, "In Progress", result.Transitions[0].To.Display)
		assert.Equal(t, "2", result.Transitions[1].ID)
		assert.Equal(t, "Close", result.Transitions[1].Display)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"ExecuteTransition",
			403,
			"forbidden",
			"Transition not allowed",
			"body",
		)

		mockAdapter.EXPECT().
			ExecuteTransition(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.executeTransition(context.Background(), executeTransitionInputDTO{
			IssueID:      "TEST-1",
			TransitionID: "invalid",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 403")
	})
}

func TestTools_AddComment(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.addComment(context.Background(), addCommentInputDTO{
			Text: "Test comment",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/text_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.addComment(context.Background(), addCommentInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "text is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentAddResponse{
			Comment: domain.TrackerComment{
				ID:     "1",
				LongID: "longid1",
				Self:   "https://api/comments/1",
				Text:   "Test comment",
			},
		}

		mockAdapter.EXPECT().
			AddComment(gomock.Any(), &domain.TrackerCommentAddRequest{
				IssueID: "TEST-1",
				Text:    "Test comment",
			}).
			Return(expectedResp, nil)

		result, err := reg.addComment(context.Background(), addCommentInputDTO{
			IssueID: "TEST-1",
			Text:    "Test comment",
		})
		require.NoError(t, err)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "Test comment", result.Text)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentAddResponse{
			Comment: domain.TrackerComment{
				ID:     "2",
				LongID: "longid2",
				Self:   "https://api/comments/2",
				Text:   "Full comment",
			},
		}

		mockAdapter.EXPECT().
			AddComment(gomock.Any(), &domain.TrackerCommentAddRequest{
				IssueID:           "TEST-1",
				Text:              "Full comment",
				AttachmentIDs:     []string{"att1", "att2"},
				MarkupType:        "md",
				Summonees:         []string{"user1", "user2"},
				MaillistSummonees: []string{"team@example.com"},
				IsAddToFollowers:  true,
			}).
			Return(expectedResp, nil)

		result, err := reg.addComment(context.Background(), addCommentInputDTO{
			IssueID:           "TEST-1",
			Text:              "Full comment",
			AttachmentIDs:     []string{"att1", "att2"},
			MarkupType:        "md",
			Summonees:         []string{"user1", "user2"},
			MaillistSummonees: []string{"team@example.com"},
			IsAddToFollowers:  true,
		})
		require.NoError(t, err)
		assert.Equal(t, "2", result.ID)
	})

	t.Run("result/maps_comment_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentAddResponse{
			Comment: domain.TrackerComment{
				ID:        "3",
				LongID:    "longid3",
				Self:      "https://api/comments/3",
				Text:      "Comment with details",
				Version:   1,
				Type:      "standard",
				Transport: "internal",
				CreatedAt: "2024-01-01T12:00:00Z",
				UpdatedAt: "2024-01-01T12:00:00Z",
				CreatedBy: &domain.TrackerUser{ID: "u1", Display: "Author"},
				UpdatedBy: &domain.TrackerUser{ID: "u1", Display: "Author"},
			},
		}

		mockAdapter.EXPECT().
			AddComment(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.addComment(context.Background(), addCommentInputDTO{
			IssueID: "TEST-1",
			Text:    "Comment with details",
		})
		require.NoError(t, err)
		assert.Equal(t, "3", result.ID)
		assert.Equal(t, "longid3", result.LongID)
		assert.Equal(t, "https://api/comments/3", result.Self)
		assert.Equal(t, "Comment with details", result.Text)
		assert.Equal(t, 1, result.Version)
		assert.Equal(t, "standard", result.Type)
		assert.Equal(t, "internal", result.Transport)
		assert.Equal(t, "2024-01-01T12:00:00Z", result.CreatedAt)
		assert.Equal(t, "2024-01-01T12:00:00Z", result.UpdatedAt)
		require.NotNil(t, result.CreatedBy)
		assert.Equal(t, "Author", result.CreatedBy.Display)
		require.NotNil(t, result.UpdatedBy)
		assert.Equal(t, "Author", result.UpdatedBy.Display)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"AddComment",
			500,
			"internal_error",
			"Failed to add comment",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			AddComment(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.addComment(context.Background(), addCommentInputDTO{
			IssueID: "TEST-1",
			Text:    "Comment",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 500")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_UpdateComment(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			CommentID: "123",
			Text:      "Updated comment",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/comment_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			IssueID: "TEST-1",
			Text:    "Updated comment",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "comment_id is required")
	})

	t.Run("validation/text_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			IssueID:   "TEST-1",
			CommentID: "123",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "text is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentUpdateResponse{
			Comment: domain.TrackerComment{
				ID:     "1",
				LongID: "longid1",
				Self:   "https://api/comments/1",
				Text:   "Updated comment",
			},
		}

		mockAdapter.EXPECT().
			UpdateComment(gomock.Any(), &domain.TrackerCommentUpdateRequest{
				IssueID:   "TEST-1",
				CommentID: "123",
				Text:      "Updated comment",
			}).
			Return(expectedResp, nil)

		result, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			IssueID:   "TEST-1",
			CommentID: "123",
			Text:      "Updated comment",
		})
		require.NoError(t, err)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "Updated comment", result.Text)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentUpdateResponse{
			Comment: domain.TrackerComment{
				ID:     "2",
				LongID: "longid2",
				Self:   "https://api/comments/2",
				Text:   "Full updated comment",
			},
		}

		mockAdapter.EXPECT().
			UpdateComment(gomock.Any(), &domain.TrackerCommentUpdateRequest{
				IssueID:           "TEST-1",
				CommentID:         "456",
				Text:              "Full updated comment",
				AttachmentIDs:     []string{"att1", "att2"},
				MarkupType:        "md",
				Summonees:         []string{"user1", "user2"},
				MaillistSummonees: []string{"team@example.com"},
			}).
			Return(expectedResp, nil)

		result, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			IssueID:           "TEST-1",
			CommentID:         "456",
			Text:              "Full updated comment",
			AttachmentIDs:     []string{"att1", "att2"},
			MarkupType:        "md",
			Summonees:         []string{"user1", "user2"},
			MaillistSummonees: []string{"team@example.com"},
		})
		require.NoError(t, err)
		assert.Equal(t, "2", result.ID)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"UpdateComment",
			500,
			"internal_error",
			"Failed to update comment",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			UpdateComment(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.updateComment(context.Background(), updateCommentInputDTO{
			IssueID:   "TEST-1",
			CommentID: "123",
			Text:      "Comment",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 500")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_DeleteComment(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteComment(context.Background(), deleteCommentInputDTO{
			CommentID: "123",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/comment_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteComment(context.Background(), deleteCommentInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "comment_id is required")
	})

	t.Run("adapter/call_and_success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			DeleteComment(gomock.Any(), &domain.TrackerCommentDeleteRequest{
				IssueID:   "TEST-1",
				CommentID: "123",
			}).
			Return(nil)

		result, err := reg.deleteComment(context.Background(), deleteCommentInputDTO{
			IssueID:   "TEST-1",
			CommentID: "123",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"DeleteComment",
			404,
			"not_found",
			"Comment not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			DeleteComment(gomock.Any(), gomock.Any()).
			Return(upstreamErr)

		_, err := reg.deleteComment(context.Background(), deleteCommentInputDTO{
			IssueID:   "TEST-1",
			CommentID: "nonexistent",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_ListAttachments(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listAttachments(context.Background(), listAttachmentsInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("adapter/call_and_returns_attachments", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedAttachments := []domain.TrackerAttachment{
			{
				ID:         "att1",
				Name:       "file1.pdf",
				ContentURL: "https://api/attachments/att1/file1.pdf",
				Mimetype:   "application/pdf",
				Size:       1024,
				CreatedAt:  "2024-01-01T12:00:00Z",
			},
			{
				ID:           "att2",
				Name:         "image.png",
				ContentURL:   "https://api/attachments/att2/image.png",
				ThumbnailURL: "https://api/attachments/att2/image_thumb.png",
				Mimetype:     "image/png",
				Size:         2048,
				CreatedAt:    "2024-01-02T12:00:00Z",
			},
		}

		mockAdapter.EXPECT().
			ListIssueAttachments(gomock.Any(), "TEST-1").
			Return(expectedAttachments, nil)

		result, err := reg.listAttachments(context.Background(), listAttachmentsInputDTO{
			IssueID: "TEST-1",
		})
		require.NoError(t, err)
		require.Len(t, result.Attachments, 2)
		assert.Equal(t, "att1", result.Attachments[0].ID)
		assert.Equal(t, "file1.pdf", result.Attachments[0].Name)
		assert.Equal(t, int64(1024), result.Attachments[0].Size)
		assert.Equal(t, "att2", result.Attachments[1].ID)
		assert.Equal(t, "image.png", result.Attachments[1].Name)
		assert.Equal(t, "https://api/attachments/att2/image_thumb.png", result.Attachments[1].ThumbnailURL)
	})

	t.Run("adapter/empty_list", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			ListIssueAttachments(gomock.Any(), "TEST-2").
			Return([]domain.TrackerAttachment{}, nil)

		result, err := reg.listAttachments(context.Background(), listAttachmentsInputDTO{
			IssueID: "TEST-2",
		})
		require.NoError(t, err)
		assert.Empty(t, result.Attachments)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"ListIssueAttachments",
			403,
			"forbidden",
			"Access denied",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			ListIssueAttachments(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.listAttachments(context.Background(), listAttachmentsInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 403")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_DeleteAttachment(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteAttachment(context.Background(), deleteAttachmentInputDTO{
			FileID: "att1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/file_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteAttachment(context.Background(), deleteAttachmentInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file_id is required")
	})

	t.Run("adapter/call_and_success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			DeleteAttachment(gomock.Any(), &domain.TrackerAttachmentDeleteRequest{
				IssueID: "TEST-1",
				FileID:  "att1",
			}).
			Return(nil)

		result, err := reg.deleteAttachment(context.Background(), deleteAttachmentInputDTO{
			IssueID: "TEST-1",
			FileID:  "att1",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"DeleteAttachment",
			404,
			"not_found",
			"Attachment not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			DeleteAttachment(gomock.Any(), gomock.Any()).
			Return(upstreamErr)

		_, err := reg.deleteAttachment(context.Background(), deleteAttachmentInputDTO{
			IssueID: "TEST-1",
			FileID:  "nonexistent",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_GetQueue(t *testing.T) {
	t.Parallel()

	t.Run("validation/queue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getQueue(context.Background(), getQueueInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue_id_or_key is required")
	})

	t.Run("adapter/call_with_expand", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedQueue := &domain.TrackerQueueDetail{
			Self:        "https://api/v3/queues/TEST",
			ID:          "1",
			Key:         "TEST",
			Name:        "Test Queue",
			Description: "Test queue description",
			Version:     1,
		}

		mockAdapter.EXPECT().
			GetQueue(gomock.Any(), "TEST", domain.TrackerGetQueueOpts{
				Expand: "all",
			}).
			Return(expectedQueue, nil)

		result, err := reg.getQueue(context.Background(), getQueueInputDTO{
			QueueID: "TEST",
			Expand:  "all",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST", result.Key)
		assert.Equal(t, "Test Queue", result.Name)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetQueue",
			404,
			"not_found",
			"Queue not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetQueue(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getQueue(context.Background(), getQueueInputDTO{
			QueueID: "NONEXISTENT",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_CreateQueue(t *testing.T) {
	t.Parallel()

	t.Run("validation/key_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Name:            "Test Queue",
			Lead:            "admin",
			DefaultType:     "task",
			DefaultPriority: "normal",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key is required")
	})

	t.Run("validation/name_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:             "TEST",
			Lead:            "admin",
			DefaultType:     "task",
			DefaultPriority: "normal",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("validation/lead_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:             "TEST",
			Name:            "Test Queue",
			DefaultType:     "task",
			DefaultPriority: "normal",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lead is required")
	})

	t.Run("validation/default_type_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:             "TEST",
			Name:            "Test Queue",
			Lead:            "admin",
			DefaultPriority: "normal",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_type is required")
	})

	t.Run("validation/default_priority_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:         "TEST",
			Name:        "Test Queue",
			Lead:        "admin",
			DefaultType: "task",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_priority is required")
	})

	t.Run("adapter/call_and_returns_queue", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerQueueCreateResponse{
			Queue: domain.TrackerQueueDetail{
				Self:    "https://api/v3/queues/TEST",
				ID:      "1",
				Key:     "TEST",
				Name:    "Test Queue",
				Version: 1,
			},
		}

		mockAdapter.EXPECT().
			CreateQueue(gomock.Any(), &domain.TrackerQueueCreateRequest{
				Key:             "TEST",
				Name:            "Test Queue",
				Lead:            "admin",
				DefaultType:     "task",
				DefaultPriority: "normal",
			}).
			Return(expectedResp, nil)

		result, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:             "TEST",
			Name:            "Test Queue",
			Lead:            "admin",
			DefaultType:     "task",
			DefaultPriority: "normal",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST", result.Key)
		assert.Equal(t, "Test Queue", result.Name)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"CreateQueue",
			400,
			"validation_error",
			"Invalid queue key",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			CreateQueue(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.createQueue(context.Background(), createQueueInputDTO{
			Key:             "invalid",
			Name:            "Test Queue",
			Lead:            "admin",
			DefaultType:     "task",
			DefaultPriority: "normal",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 400")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_DeleteQueue(t *testing.T) {
	t.Parallel()

	t.Run("validation/queue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteQueue(context.Background(), deleteQueueInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue_id_or_key is required")
	})

	t.Run("adapter/call_and_success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			DeleteQueue(gomock.Any(), &domain.TrackerQueueDeleteRequest{
				QueueID: "TEST",
			}).
			Return(nil)

		result, err := reg.deleteQueue(context.Background(), deleteQueueInputDTO{
			QueueID: "TEST",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"DeleteQueue",
			403,
			"forbidden",
			"Not authorized",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			DeleteQueue(gomock.Any(), gomock.Any()).
			Return(upstreamErr)

		_, err := reg.deleteQueue(context.Background(), deleteQueueInputDTO{
			QueueID: "TEST",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 403")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_RestoreQueue(t *testing.T) {
	t.Parallel()

	t.Run("validation/queue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.restoreQueue(context.Background(), restoreQueueInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue_id_or_key is required")
	})

	t.Run("adapter/call_and_returns_queue", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerQueueRestoreResponse{
			Queue: domain.TrackerQueueDetail{
				Self:    "https://api/v3/queues/TEST",
				ID:      "1",
				Key:     "TEST",
				Name:    "Test Queue",
				Version: 2,
			},
		}

		mockAdapter.EXPECT().
			RestoreQueue(gomock.Any(), &domain.TrackerQueueRestoreRequest{
				QueueID: "TEST",
			}).
			Return(expectedResp, nil)

		result, err := reg.restoreQueue(context.Background(), restoreQueueInputDTO{
			QueueID: "TEST",
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST", result.Key)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"RestoreQueue",
			404,
			"not_found",
			"Queue not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			RestoreQueue(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.restoreQueue(context.Background(), restoreQueueInputDTO{
			QueueID: "NONEXISTENT",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_GetCurrentUser(t *testing.T) {
	t.Parallel()

	t.Run("adapter/call_and_returns_user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedUser := &domain.TrackerUserDetail{
			Self:       "https://api/v3/users/1",
			ID:         "1",
			UID:        "123456",
			Login:      "testuser",
			Display:    "Test User",
			FirstName:  "Test",
			LastName:   "User",
			Email:      "test@example.com",
			HasLicense: true,
			Dismissed:  false,
			External:   false,
		}

		mockAdapter.EXPECT().
			GetCurrentUser(gomock.Any()).
			Return(expectedUser, nil)

		result, err := reg.getCurrentUser(context.Background(), getCurrentUserInputDTO{})
		require.NoError(t, err)
		assert.Equal(t, "testuser", result.Login)
		assert.Equal(t, "Test User", result.Display)
		assert.Equal(t, "123456", result.UID)
		assert.True(t, result.HasLicense)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetCurrentUser",
			401,
			"unauthorized",
			"Not authorized",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetCurrentUser(gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getCurrentUser(context.Background(), getCurrentUserInputDTO{})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 401")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_ListUsers(t *testing.T) {
	t.Parallel()

	t.Run("validation/per_page_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listUsers(context.Background(), listUsersInputDTO{PerPage: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("validation/page_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listUsers(context.Background(), listUsersInputDTO{Page: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be non-negative")
	})

	t.Run("adapter/call_with_pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResult := &domain.TrackerUsersPage{
			Users: []domain.TrackerUserDetail{
				{
					ID:      "1",
					Login:   "user1",
					Display: "User One",
				},
				{
					ID:      "2",
					Login:   "user2",
					Display: "User Two",
				},
			},
			TotalCount: 100,
			TotalPages: 10,
		}

		mockAdapter.EXPECT().
			ListUsers(gomock.Any(), domain.TrackerListUsersOpts{
				PerPage: 10,
				Page:    2,
			}).
			Return(expectedResult, nil)

		result, err := reg.listUsers(context.Background(), listUsersInputDTO{
			PerPage: 10,
			Page:    2,
		})
		require.NoError(t, err)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, "user1", result.Users[0].Login)
		assert.Equal(t, 100, result.TotalCount)
		assert.Equal(t, 10, result.TotalPages)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"ListUsers",
			500,
			"internal_error",
			"Internal server error",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			ListUsers(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.listUsers(context.Background(), listUsersInputDTO{})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 500")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_GetUser(t *testing.T) {
	t.Parallel()

	t.Run("validation/user_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getUser(context.Background(), getUserInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user_id is required")
	})

	t.Run("adapter/call_and_returns_user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedUser := &domain.TrackerUserDetail{
			Self:    "https://api.tracker/v3/users/testuser",
			ID:      "1",
			UID:     "123456",
			Login:   "testuser",
			Display: "Test User",
		}

		mockAdapter.EXPECT().
			GetUser(gomock.Any(), "testuser").
			Return(expectedUser, nil)

		result, err := reg.getUser(context.Background(), getUserInputDTO{
			UserID: "testuser",
		})
		require.NoError(t, err)
		assert.Equal(t, "testuser", result.Login)
		assert.Equal(t, "Test User", result.Display)
		assert.Equal(t, "123456", result.UID)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetUser",
			404,
			"not_found",
			"User not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetUser(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getUser(context.Background(), getUserInputDTO{
			UserID: "nonexistent",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_ListLinks(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listLinks(context.Background(), listLinksInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("adapter/call_and_returns_links", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedLinks := []domain.TrackerLink{
			{
				ID:        "link1",
				Self:      "https://api/v3/issues/TEST-1/links/link1",
				Direction: "outward",
				Type: &domain.TrackerLinkType{
					ID:      "relates",
					Inward:  "is related to",
					Outward: "relates to",
				},
				Object: &domain.TrackerLinkedIssue{
					Self:    "https://api/v3/issues/TEST-2",
					ID:      "2",
					Key:     "TEST-2",
					Display: "TEST-2: Second issue",
				},
			},
		}

		mockAdapter.EXPECT().
			ListIssueLinks(gomock.Any(), "TEST-1").
			Return(expectedLinks, nil)

		result, err := reg.listLinks(context.Background(), listLinksInputDTO{
			IssueID: "TEST-1",
		})
		require.NoError(t, err)
		require.Len(t, result.Links, 1)
		assert.Equal(t, "link1", result.Links[0].ID)
		assert.Equal(t, "outward", result.Links[0].Direction)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"ListIssueLinks",
			404,
			"not_found",
			"Issue not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			ListIssueLinks(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.listLinks(context.Background(), listLinksInputDTO{
			IssueID: "NONEXISTENT",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_CreateLink(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createLink(context.Background(), createLinkInputDTO{
			Relationship: "relates",
			TargetIssue:  "TEST-2",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/relationship_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createLink(context.Background(), createLinkInputDTO{
			IssueID:     "TEST-1",
			TargetIssue: "TEST-2",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "relationship is required")
	})

	t.Run("validation/target_issue_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.createLink(context.Background(), createLinkInputDTO{
			IssueID:      "TEST-1",
			Relationship: "relates",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target_issue is required")
	})

	t.Run("adapter/call_and_returns_link", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedLink := &domain.TrackerLinkCreateResponse{
			Link: domain.TrackerLink{
				ID:        "newlink",
				Self:      "https://api/v3/issues/TEST-1/links/newlink",
				Direction: "outward",
				Type: &domain.TrackerLinkType{
					ID:      "relates",
					Inward:  "is related to",
					Outward: "relates to",
				},
				Object: &domain.TrackerLinkedIssue{
					Self: "https://api/v3/issues/TEST-2",
					Key:  "TEST-2",
				},
			},
		}

		mockAdapter.EXPECT().
			CreateLink(gomock.Any(), &domain.TrackerLinkCreateRequest{
				IssueID:      "TEST-1",
				Relationship: "relates",
				TargetIssue:  "TEST-2",
			}).
			Return(expectedLink, nil)

		result, err := reg.createLink(context.Background(), createLinkInputDTO{
			IssueID:      "TEST-1",
			Relationship: "relates",
			TargetIssue:  "TEST-2",
		})
		require.NoError(t, err)
		assert.Equal(t, "newlink", result.ID)
		assert.Equal(t, "relates", result.Type.ID)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"CreateLink",
			400,
			"validation_error",
			"Invalid relationship",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			CreateLink(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.createLink(context.Background(), createLinkInputDTO{
			IssueID:      "TEST-1",
			Relationship: "invalid",
			TargetIssue:  "TEST-2",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 400")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_DeleteLink(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteLink(context.Background(), deleteLinkInputDTO{
			LinkID: "link1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/link_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.deleteLink(context.Background(), deleteLinkInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "link_id is required")
	})

	t.Run("adapter/call_and_success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		mockAdapter.EXPECT().
			DeleteLink(gomock.Any(), &domain.TrackerLinkDeleteRequest{
				IssueID: "TEST-1",
				LinkID:  "link1",
			}).
			Return(nil)

		result, err := reg.deleteLink(context.Background(), deleteLinkInputDTO{
			IssueID: "TEST-1",
			LinkID:  "link1",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"DeleteLink",
			404,
			"not_found",
			"Link not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			DeleteLink(gomock.Any(), gomock.Any()).
			Return(upstreamErr)

		_, err := reg.deleteLink(context.Background(), deleteLinkInputDTO{
			IssueID: "TEST-1",
			LinkID:  "nonexistent",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_GetChangelog(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getChangelog(context.Background(), getChangelogInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/per_page_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getChangelog(context.Background(), getChangelogInputDTO{
			IssueID: "TEST-1",
			PerPage: -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("adapter/call_with_per_page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedEntries := []domain.TrackerChangelogEntry{
			{
				ID:        "entry1",
				Self:      "https://api/v3/issues/TEST-1/changelog/entry1",
				UpdatedAt: "2024-01-01T10:00:00.000+0000",
				Type:      "IssueUpdated",
				Fields: []domain.TrackerChangelogFieldChange{
					{Field: "status", From: "open", To: "inProgress"},
				},
			},
		}

		mockAdapter.EXPECT().
			GetIssueChangelog(gomock.Any(), "TEST-1", domain.TrackerGetChangelogOpts{
				PerPage: 100,
			}).
			Return(expectedEntries, nil)

		result, err := reg.getChangelog(context.Background(), getChangelogInputDTO{
			IssueID: "TEST-1",
			PerPage: 100,
		})
		require.NoError(t, err)
		require.Len(t, result.Entries, 1)
		assert.Equal(t, "entry1", result.Entries[0].ID)
		assert.Equal(t, "IssueUpdated", result.Entries[0].Type)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetIssueChangelog",
			404,
			"not_found",
			"Issue not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetIssueChangelog(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getChangelog(context.Background(), getChangelogInputDTO{
			IssueID: "NONEXISTENT",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_MoveIssue(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.moveIssue(context.Background(), moveIssueInputDTO{
			Queue: "NEWQUEUE",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/queue_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.moveIssue(context.Background(), moveIssueInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue is required")
	})

	t.Run("adapter/call_with_initial_status", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerIssueMoveResponse{
			Issue: domain.TrackerIssue{
				Self:    "https://api/v3/issues/NEWQUEUE-1",
				ID:      "12345",
				Key:     "NEWQUEUE-1",
				Summary: "Test Issue",
			},
		}

		mockAdapter.EXPECT().
			MoveIssue(gomock.Any(), &domain.TrackerIssueMoveRequest{
				IssueID:       "TEST-1",
				Queue:         "NEWQUEUE",
				InitialStatus: true,
			}).
			Return(expectedResp, nil)

		result, err := reg.moveIssue(context.Background(), moveIssueInputDTO{
			IssueID:       "TEST-1",
			Queue:         "NEWQUEUE",
			InitialStatus: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "NEWQUEUE-1", result.Key)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"MoveIssue",
			403,
			"forbidden",
			"No permission to move issue",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			MoveIssue(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.moveIssue(context.Background(), moveIssueInputDTO{
			IssueID: "TEST-1",
			Queue:   "NEWQUEUE",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 403")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_ListProjectComments(t *testing.T) {
	t.Parallel()

	t.Run("validation/project_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.listProjectComments(context.Background(), listProjectCommentsInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project_id is required")
	})

	t.Run("adapter/call_with_expand", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedComments := []domain.TrackerProjectComment{
			{
				ID:        "1",
				LongID:    "longid1",
				Self:      "https://api/v3/entities/project/123/comments/1",
				Text:      "Project comment",
				CreatedAt: "2024-01-01T10:00:00.000+0000",
			},
		}

		mockAdapter.EXPECT().
			ListProjectComments(gomock.Any(), "123", domain.TrackerListProjectCommentsOpts{
				Expand: "all",
			}).
			Return(expectedComments, nil)

		result, err := reg.listProjectComments(context.Background(), listProjectCommentsInputDTO{
			ProjectID: "123",
			Expand:    "all",
		})
		require.NoError(t, err)
		require.Len(t, result.Comments, 1)
		assert.Equal(t, "1", result.Comments[0].ID)
		assert.Equal(t, "Project comment", result.Comments[0].Text)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"ListProjectComments",
			404,
			"not_found",
			"Project not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			ListProjectComments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.listProjectComments(context.Background(), listProjectCommentsInputDTO{
			ProjectID: "nonexistent",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
		assert.NotContains(t, errStr, "secrets")
	})
}
