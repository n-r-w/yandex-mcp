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

		_, err := reg.GetIssue(context.Background(), GetIssueInput{})
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

		result, err := reg.GetIssue(context.Background(), GetIssueInput{
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

		_, err := reg.GetIssue(context.Background(), GetIssueInput{IssueID: "MISSING-1"})
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

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{PerPage: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("returns error when page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{Page: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be non-negative")
	})

	t.Run("returns error when per_scroll is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{PerScroll: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_scroll must be non-negative")
	})

	t.Run("returns error when per_scroll exceeds max", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{PerScroll: 1001})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_scroll must not exceed 1000")
	})

	t.Run("returns error when scroll_ttl_millis is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{ScrollTTLMillis: -1})
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

		input := SearchIssuesInput{
			Filter:          map[string]any{"status": "open"},
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

		result, err := reg.SearchIssues(context.Background(), input)
		require.NoError(t, err)
		assert.Len(t, result.Issues, 2)
		assert.Equal(t, 100, result.TotalCount)
		assert.Equal(t, 10, result.TotalPages)
		assert.Equal(t, "scroll123", result.ScrollID)
		assert.Equal(t, "token456", result.ScrollToken)
		assert.Equal(t, "https://api/next", result.NextLink)
	})

	t.Run("returns error when filter contains non-string value", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.SearchIssues(context.Background(), SearchIssuesInput{
			Filter: map[string]any{"status": 123},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "filter value for key \"status\" must be a string")
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

		result, err := reg.CountIssues(context.Background(), CountIssuesInput{
			Filter: map[string]any{"assignee": "me"},
			Query:  "Queue: PROJ",
		})
		require.NoError(t, err)
		assert.Equal(t, 42, result.Count)
	})

	t.Run("returns error when filter contains non-string value", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.CountIssues(context.Background(), CountIssuesInput{
			Filter: map[string]any{"priority": []int{1, 2, 3}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "filter value for key \"priority\" must be a string")
	})
}

func TestTools_ListTransitions(t *testing.T) {
	t.Parallel()

	t.Run("returns error when issue_id_or_key is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.ListTransitions(context.Background(), ListTransitionsInput{})
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

		result, err := reg.ListTransitions(context.Background(), ListTransitionsInput{
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

		_, err := reg.ListQueues(context.Background(), ListQueuesInput{PerPage: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("returns error when page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.ListQueues(context.Background(), ListQueuesInput{Page: -1})
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

		result, err := reg.ListQueues(context.Background(), ListQueuesInput{
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

		_, err := reg.ListComments(context.Background(), ListCommentsInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("returns error when per_page is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.ListComments(context.Background(), ListCommentsInput{
			IssueID: "TEST-1",
			PerPage: -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be non-negative")
	})

	t.Run("returns error when id is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.ListComments(context.Background(), ListCommentsInput{
			IssueID: "TEST-1",
			ID:      -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id must be non-negative")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResult := &domain.TrackerCommentsPage{
			Comments: []domain.TrackerComment{
				{
					ID:        1,
					LongID:    "longid1",
					Text:      "First comment",
					CreatedBy: &domain.TrackerUser{ID: "user1", Display: "User One"},
				},
				{
					ID:        2,
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
				ID:      100,
			}).
			Return(expectedResult, nil)

		result, err := reg.ListComments(context.Background(), ListCommentsInput{
			IssueID: "TEST-1",
			Expand:  "html",
			PerPage: 20,
			ID:      100,
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

		_, err := reg.GetIssue(context.Background(), GetIssueInput{IssueID: "TEST-1"})
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

		_, err := reg.GetIssue(context.Background(), GetIssueInput{IssueID: "TEST-1"})
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

		result, err := reg.GetIssue(context.Background(), GetIssueInput{IssueID: "PROJ-1"})
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

		_, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		_, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		result, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		result, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		result, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		_, err := reg.CreateIssue(context.Background(), CreateIssueInput{
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

		_, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
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

		_, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
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

		result, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
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

		result, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
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

		result, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
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

		_, err := reg.UpdateIssue(context.Background(), UpdateIssueInput{
			IssueID: "MISSING-1",
			Summary: "Updated",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 404")
	})
}

func TestTools_ExecuteTransition(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		_, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		result, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		result, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		result, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		_, err := reg.ExecuteTransition(context.Background(), ExecuteTransitionInput{
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

		_, err := reg.AddComment(context.Background(), AddCommentInput{
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

		_, err := reg.AddComment(context.Background(), AddCommentInput{
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
				ID:     1,
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

		result, err := reg.AddComment(context.Background(), AddCommentInput{
			IssueID: "TEST-1",
			Text:    "Test comment",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "Test comment", result.Text)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentAddResponse{
			Comment: domain.TrackerComment{
				ID:     2,
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
				MarkupType:        "wiki",
				Summonees:         []string{"user1", "user2"},
				MaillistSummonees: []string{"team@example.com"},
				IsAddToFollowers:  true,
			}).
			Return(expectedResp, nil)

		result, err := reg.AddComment(context.Background(), AddCommentInput{
			IssueID:           "TEST-1",
			Text:              "Full comment",
			AttachmentIDs:     []string{"att1", "att2"},
			MarkupType:        "wiki",
			Summonees:         []string{"user1", "user2"},
			MaillistSummonees: []string{"team@example.com"},
			IsAddToFollowers:  true,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.ID)
	})

	t.Run("result/maps_comment_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		expectedResp := &domain.TrackerCommentAddResponse{
			Comment: domain.TrackerComment{
				ID:        3,
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

		result, err := reg.AddComment(context.Background(), AddCommentInput{
			IssueID: "TEST-1",
			Text:    "Comment with details",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.ID)
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

		_, err := reg.AddComment(context.Background(), AddCommentInput{
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
