//nolint:exhaustruct // test file uses partial struct initialization for clarity
package tracker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestTools_GetAttachment(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/attachment_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attachment_id is required")
	})

	t.Run("validation/file_name_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file_name is required")
	})

	t.Run("validation/save_path_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			FileName:     "attachment.txt",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "save_path is required")
	})

	t.Run("adapter/call_and_returns_content", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())
		reg.baseDir = t.TempDir()
		savePath := filepath.Join("attachments", "attachment.txt")

		payload := []byte("hello world")
		expected := &domain.TrackerAttachmentContent{
			FileName:    "attachment.txt",
			ContentType: "text/plain",
			Data:        payload,
		}

		mockAdapter.EXPECT().
			GetIssueAttachment(gomock.Any(), "TEST-1", "4159", "attachment.txt").
			Return(expected, nil)

		result, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			FileName:     "attachment.txt",
			SavePath:     savePath,
		})
		require.NoError(t, err)
		assert.Equal(t, "attachment.txt", result.FileName)
		assert.Equal(t, "text/plain", result.ContentType)
		assert.Equal(t, int64(len(payload)), result.Size)
		assert.Equal(t, savePath, result.SavedPath)

		stored, err := os.ReadFile(filepath.Join(reg.baseDir, savePath))
		require.NoError(t, err)
		assert.Equal(t, payload, stored)
	})

	t.Run("validation/save_path_exists_without_override", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())
		reg.baseDir = t.TempDir()
		savePath := filepath.Join("attachments", "existing.txt")
		fullPath := filepath.Join(reg.baseDir, savePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte("data"), 0o644))

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			FileName:     "attachment.txt",
			SavePath:     savePath,
			Override:     false,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "save_path already exists")
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetIssueAttachment",
			403,
			"forbidden",
			"Access denied",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetIssueAttachment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getAttachment(context.Background(), getAttachmentInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			FileName:     "attachment.txt",
			SavePath:     filepath.Join("attachments", "attachment.txt"),
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceTracker)
		assert.Contains(t, errStr, "HTTP 403")
		assert.NotContains(t, errStr, "secrets")
	})
}

func TestTools_GetAttachmentPreview(t *testing.T) {
	t.Parallel()

	t.Run("validation/issue_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachmentPreview(context.Background(), getAttachmentPreviewInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_id_or_key is required")
	})

	t.Run("validation/attachment_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachmentPreview(context.Background(), getAttachmentPreviewInputDTO{
			IssueID: "TEST-1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attachment_id is required")
	})

	t.Run("validation/save_path_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		_, err := reg.getAttachmentPreview(context.Background(), getAttachmentPreviewInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "save_path is required")
	})

	t.Run("adapter/call_and_returns_preview", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())
		reg.baseDir = t.TempDir()
		savePath := filepath.Join("attachments", "preview.png")

		payload := []byte{0x1, 0x2, 0x3}
		expected := &domain.TrackerAttachmentContent{
			ContentType: "image/png",
			Data:        payload,
		}

		mockAdapter.EXPECT().
			GetIssueAttachmentPreview(gomock.Any(), "TEST-1", "4159").
			Return(expected, nil)

		result, err := reg.getAttachmentPreview(context.Background(), getAttachmentPreviewInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			SavePath:     savePath,
		})
		require.NoError(t, err)
		assert.Equal(t, "image/png", result.ContentType)
		assert.Equal(t, int64(len(payload)), result.Size)
		assert.Equal(t, savePath, result.SavedPath)

		stored, err := os.ReadFile(filepath.Join(reg.baseDir, savePath))
		require.NoError(t, err)
		assert.Equal(t, payload, stored)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockITrackerAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.TrackerAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceTracker,
			"GetIssueAttachmentPreview",
			404,
			"not_found",
			"Attachment not found",
			"body with secrets",
		)

		mockAdapter.EXPECT().
			GetIssueAttachmentPreview(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.getAttachmentPreview(context.Background(), getAttachmentPreviewInputDTO{
			IssueID:      "TEST-1",
			AttachmentID: "4159",
			SavePath:     filepath.Join("attachments", "preview.png"),
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
