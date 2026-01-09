//nolint:exhaustruct // test file uses partial struct initialization for clarity
package wiki

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

func TestTools_GetPageBySlug(t *testing.T) {
	t.Parallel()

	t.Run("returns error when slug is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.GetPageBySlug(context.Background(), GetPageBySlugInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slug is required")
	})

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:       123,
			PageType: "doc",
			Slug:     "test/page",
			Title:    "Test Page",
			Content:  "Hello",
		}

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "test/page", domain.WikiGetPageOpts{Fields: []string{"content", "attributes"}}).
			Return(expectedPage, nil)

		input := GetPageBySlugInput{
			Slug:   "test/page",
			Fields: []string{"content", "attributes"},
		}

		result, err := reg.GetPageBySlug(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "Test Page", result.Title)
		assert.Equal(t, "test/page", result.Slug)
	})

	t.Run("returns safe error on upstream error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.UpstreamError{
			Service:    domain.ServiceWiki,
			Operation:  "GetPageBySlug",
			HTTPStatus: 404,
			Message:    "Page not found",
		}

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "missing", domain.WikiGetPageOpts{}).
			Return(nil, upstreamErr)

		_, err := reg.GetPageBySlug(context.Background(), GetPageBySlugInput{Slug: "missing"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wiki")
		assert.Contains(t, err.Error(), "GetPageBySlug")
		assert.Contains(t, err.Error(), "HTTP 404")
		assert.NotContains(t, err.Error(), "token")
	})
}

func TestTools_GetPageByID(t *testing.T) {
	t.Parallel()

	t.Run("returns error when id is not positive", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.GetPageByID(context.Background(), GetPageByIDInput{PageID: 0})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")

		_, err = reg.GetPageByID(context.Background(), GetPageByIDInput{PageID: -1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:       456,
			PageType: "doc",
			Slug:     "another/page",
			Title:    "Another Page",
		}

		mockAdapter.EXPECT().
			GetPageByID(gomock.Any(), int64(456), domain.WikiGetPageOpts{Fields: []string{"title"}}).
			Return(expectedPage, nil)

		result, err := reg.GetPageByID(context.Background(), GetPageByIDInput{
			PageID: 456,
			Fields: []string{"title"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(456), result.ID)
		assert.Equal(t, "Another Page", result.Title)
	})

	t.Run("maps attributes correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:    789,
			Title: "With Attrs",
			Attributes: &domain.WikiAttributes{
				CommentsCount:   5,
				CommentsEnabled: true,
				CreatedAt:       "2024-01-01T00:00:00Z",
				Lang:            "en",
			},
		}

		mockAdapter.EXPECT().
			GetPageByID(gomock.Any(), int64(789), domain.WikiGetPageOpts{}).
			Return(expectedPage, nil)

		result, err := reg.GetPageByID(context.Background(), GetPageByIDInput{PageID: 789})
		require.NoError(t, err)
		require.NotNil(t, result.Attributes)
		assert.Equal(t, 5, result.Attributes.CommentsCount)
		assert.True(t, result.Attributes.CommentsEnabled)
		assert.Equal(t, "en", result.Attributes.Lang)
	})
}

func TestTools_ListResources(t *testing.T) {
	t.Parallel()

	t.Run("returns error when page_id is not positive", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListResources(context.Background(), ListResourcesInput{PageID: 0})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("returns error when page_size is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListResources(context.Background(), ListResourcesInput{
			PageID:   123,
			PageSize: -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_size must be non-negative")
	})

	t.Run("returns error when page_size exceeds max", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListResources(context.Background(), ListResourcesInput{
			PageID:   123,
			PageSize: 51,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_size must not exceed 50")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResult := &domain.WikiResourcesPage{
			Resources: []domain.WikiResource{
				{
					Type: "attachment",
					Attachment: &domain.WikiAttachment{
						Name: "file.pdf",
					},
				},
			},
			NextCursor: "next123",
			PrevCursor: "prev123",
		}

		mockAdapter.EXPECT().
			ListPageResources(gomock.Any(), int64(100), domain.WikiListResourcesOpts{
				Cursor:         "cursor1",
				PageSize:       20,
				OrderBy:        "created_at",
				OrderDirection: "desc",
				Query:          "test",
				Types:          "attachment",
			}).
			Return(expectedResult, nil)

		input := ListResourcesInput{
			PageID:         100,
			Cursor:         "cursor1",
			PageSize:       20,
			OrderBy:        "created_at",
			OrderDirection: "desc",
			Q:              "test",
			Types:          "attachment",
		}

		result, err := reg.ListResources(context.Background(), input)
		require.NoError(t, err)
		assert.Len(t, result.Resources, 1)
		assert.Equal(t, "next123", result.NextCursor)
		assert.Equal(t, "prev123", result.PrevCursor)
	})

	t.Run("maps attachment resource correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResult := &domain.WikiResourcesPage{
			Resources: []domain.WikiResource{
				{
					Type: "attachment",
					Attachment: &domain.WikiAttachment{
						ID:          101,
						Name:        "document.pdf",
						Size:        1024,
						MIMEType:    "application/pdf",
						DownloadURL: "https://example.com/download",
						CreatedAt:   "2024-01-01T00:00:00Z",
						HasPreview:  true,
					},
				},
			},
		}

		mockAdapter.EXPECT().
			ListPageResources(gomock.Any(), int64(100), gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.ListResources(context.Background(), ListResourcesInput{PageID: 100})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "attachment", res.Type)
		require.NotNil(t, res.Item)

		attachment, ok := res.Item.(AttachmentOutput)
		require.True(t, ok, "expected AttachmentOutput, got %T", res.Item)
		assert.Equal(t, int64(101), attachment.ID)
		assert.Equal(t, "document.pdf", attachment.Name)
		assert.Equal(t, int64(1024), attachment.Size)
		assert.Equal(t, "application/pdf", attachment.Mimetype)
		assert.Equal(t, "https://example.com/download", attachment.DownloadURL)
		assert.Equal(t, "2024-01-01T00:00:00Z", attachment.CreatedAt)
		assert.True(t, attachment.HasPreview)
	})

	t.Run("maps sharepoint resource correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResult := &domain.WikiResourcesPage{
			Resources: []domain.WikiResource{
				{
					Type: "sharepoint_resource",
					Sharepoint: &domain.WikiSharepointResource{
						ID:        202,
						Title:     "Important Document",
						Doctype:   "docx",
						CreatedAt: "2024-02-01T10:00:00Z",
					},
				},
			},
		}

		mockAdapter.EXPECT().
			ListPageResources(gomock.Any(), int64(100), gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.ListResources(context.Background(), ListResourcesInput{PageID: 100})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "sharepoint_resource", res.Type)
		require.NotNil(t, res.Item)

		sharepoint, ok := res.Item.(SharepointResourceOutput)
		require.True(t, ok, "expected SharepointResourceOutput, got %T", res.Item)
		assert.Equal(t, int64(202), sharepoint.ID)
		assert.Equal(t, "Important Document", sharepoint.Title)
		assert.Equal(t, "docx", sharepoint.Doctype)
		assert.Equal(t, "2024-02-01T10:00:00Z", sharepoint.CreatedAt)
	})

	t.Run("maps grid resource correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResult := &domain.WikiResourcesPage{
			Resources: []domain.WikiResource{
				{
					Type: "grid",
					Grid: &domain.WikiGridResource{
						ID:        "grid-xyz-123",
						Title:     "Sales Data",
						CreatedAt: "2024-03-01T15:30:00Z",
					},
				},
			},
		}

		mockAdapter.EXPECT().
			ListPageResources(gomock.Any(), int64(100), gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.ListResources(context.Background(), ListResourcesInput{PageID: 100})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "grid", res.Type)
		require.NotNil(t, res.Item)

		grid, ok := res.Item.(GridResourceOutput)
		require.True(t, ok, "expected GridResourceOutput, got %T", res.Item)
		assert.Equal(t, "grid-xyz-123", grid.ID)
		assert.Equal(t, "Sales Data", grid.Title)
		assert.Equal(t, "2024-03-01T15:30:00Z", grid.CreatedAt)
	})
}

func TestTools_ListGrids(t *testing.T) {
	t.Parallel()

	t.Run("returns error when page_id is not positive", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListGrids(context.Background(), ListGridsInput{PageID: 0})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("returns error when page_size is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListGrids(context.Background(), ListGridsInput{
			PageID:   123,
			PageSize: -1,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_size must be non-negative")
	})

	t.Run("returns error when page_size exceeds max", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.ListGrids(context.Background(), ListGridsInput{
			PageID:   123,
			PageSize: 51,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_size must not exceed 50")
	})

	t.Run("calls adapter with correct parameters and maps pagination", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResult := &domain.WikiGridsPage{
			Grids: []domain.WikiGridSummary{
				{ID: "grid1", Title: "Grid 1", CreatedAt: "2024-01-01"},
				{ID: "grid2", Title: "Grid 2", CreatedAt: "2024-01-02"},
			},
			NextCursor: "next-grid",
		}

		mockAdapter.EXPECT().
			ListPageGrids(gomock.Any(), int64(200), domain.WikiListGridsOpts{
				Cursor:   "cur",
				PageSize: 10,
			}).
			Return(expectedResult, nil)

		result, err := reg.ListGrids(context.Background(), ListGridsInput{
			PageID:   200,
			Cursor:   "cur",
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, result.Grids, 2)
		assert.Equal(t, "grid1", result.Grids[0].ID)
		assert.Equal(t, "next-grid", result.NextCursor)
	})
}

func TestTools_GetGrid(t *testing.T) {
	t.Parallel()

	t.Run("returns error when grid_id is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.GetGrid(context.Background(), GetGridInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "grid_id is required")
	})

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedGrid := &domain.WikiGrid{
			ID:       "grid123",
			Title:    "My Grid",
			Revision: "5",
			Structure: []domain.WikiColumn{
				{Slug: "col1", Title: "Column 1", Type: "string"},
			},
			Rows: []domain.WikiGridRow{
				{ID: "row1", Cells: map[string]domain.WikiGridCell{"col1": {Value: "value1"}}},
			},
		}

		mockAdapter.EXPECT().
			GetGridByID(gomock.Any(), "grid123", domain.WikiGetGridOpts{
				Fields:   []string{"rows"},
				Filter:   "col1 = 'test'",
				OnlyCols: "col1",
				Revision: 5,
			}).
			Return(expectedGrid, nil)

		result, err := reg.GetGrid(context.Background(), GetGridInput{
			GridID:   "grid123",
			Fields:   []string{"rows"},
			Filter:   "col1 = 'test'",
			OnlyCols: "col1",
			Revision: 5,
		})
		require.NoError(t, err)
		assert.Equal(t, "grid123", result.ID)
		assert.Equal(t, "My Grid", result.Title)
		assert.Equal(t, "5", result.Revision)
		assert.Len(t, result.Structure, 1)
		assert.Len(t, result.Rows, 1)
	})

	t.Run("maps grid attributes correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedGrid := &domain.WikiGrid{
			ID:    "gridWithAttrs",
			Title: "Grid With Attrs",
			Attributes: &domain.WikiAttributes{
				CreatedAt:  "2024-01-01T00:00:00Z",
				ModifiedAt: "2024-02-01T00:00:00Z",
				IsReadonly: true,
			},
		}

		mockAdapter.EXPECT().
			GetGridByID(gomock.Any(), "gridWithAttrs", domain.WikiGetGridOpts{}).
			Return(expectedGrid, nil)

		result, err := reg.GetGrid(context.Background(), GetGridInput{GridID: "gridWithAttrs"})
		require.NoError(t, err)
		require.NotNil(t, result.Attributes)
		assert.Equal(t, "2024-01-01T00:00:00Z", result.Attributes.CreatedAt)
		assert.True(t, result.Attributes.IsReadonly)
	})

	t.Run("maps grid cell values as strings", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedGrid := &domain.WikiGrid{
			ID:       "gridCellTypes",
			Title:    "Grid Cell Types Test",
			Revision: "1",
			Structure: []domain.WikiColumn{
				{Slug: "col_string", Title: "String Column", Type: "string"},
				{Slug: "col_number", Title: "Number Column", Type: "number"},
				{Slug: "col_bool", Title: "Boolean Column", Type: "boolean"},
			},
			Rows: []domain.WikiGridRow{
				{
					ID: "row1",
					Cells: map[string]domain.WikiGridCell{
						"col_string": {Value: "hello"},
						"col_number": {Value: "42.5"},
						"col_bool":   {Value: "true"},
					},
				},
				{
					ID: "row2",
					Cells: map[string]domain.WikiGridCell{
						"col_string": {Value: "world"},
						"col_number": {Value: "0"},
						"col_bool":   {Value: "false"},
					},
				},
			},
		}

		mockAdapter.EXPECT().
			GetGridByID(gomock.Any(), "gridCellTypes", domain.WikiGetGridOpts{}).
			Return(expectedGrid, nil)

		result, err := reg.GetGrid(context.Background(), GetGridInput{GridID: "gridCellTypes"})
		require.NoError(t, err)
		require.Len(t, result.Rows, 2)

		row1 := result.Rows[0]
		assert.Equal(t, "row1", row1.ID)
		require.Len(t, row1.Cells, 3)

		val1, ok := row1.Cells["col_string"]
		require.True(t, ok)
		assert.Equal(t, "hello", val1)
		assert.IsType(t, "", val1, "col_string value should be string")

		val2, ok := row1.Cells["col_number"]
		require.True(t, ok)
		assert.Equal(t, "42.5", val2)
		assert.IsType(t, "", val2, "col_number value should be string (stringified)")

		val3, ok := row1.Cells["col_bool"]
		require.True(t, ok)
		assert.Equal(t, "true", val3)
		assert.IsType(t, "", val3, "col_bool value should be string (stringified)")

		row2 := result.Rows[1]
		assert.Equal(t, "row2", row2.ID)

		val4, ok := row2.Cells["col_number"]
		require.True(t, ok)
		assert.Equal(t, "0", val4)
		assert.IsType(t, "", val4, "col_number value should be string (stringified)")

		val5, ok := row2.Cells["col_bool"]
		require.True(t, ok)
		assert.Equal(t, "false", val5)
		assert.IsType(t, "", val5, "col_bool value should be string (stringified)")
	})
}

func TestTools_ErrorShaping(t *testing.T) {
	t.Parallel()

	t.Run("upstream error is shaped safely", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"GetPageBySlug",
			500,
			"internal_error",
			"Internal server error",
			"detailed body with token: Bearer xyz123",
		)

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "test", domain.WikiGetPageOpts{}).
			Return(nil, upstreamErr)

		_, err := reg.GetPageBySlug(context.Background(), GetPageBySlugInput{Slug: "test"})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 500")
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "xyz123")
	})

	t.Run("non-upstream error is shaped safely", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		// Simulate an error that contains sensitive data
		sensitiveErr := errors.New("connection failed: Authorization header: Bearer secret-token-123")

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "test", domain.WikiGetPageOpts{}).
			Return(nil, sensitiveErr)

		_, err := reg.GetPageBySlug(context.Background(), GetPageBySlugInput{Slug: "test"})
		require.Error(t, err)
		errStr := err.Error()
		// Non-upstream errors should return a generic safe message
		assert.Equal(t, "wiki: internal error", errStr)
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret-token-123")
	})
}

func TestTools_CreatePage(t *testing.T) {
	t.Parallel()

	t.Run("validation/slug_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreatePage(context.Background(), CreatePageInput{
			Title:    "Test Page",
			PageType: "page",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slug is required")
	})

	t.Run("validation/title_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:     "test/page",
			PageType: "page",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("validation/page_type_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:  "test/page",
			Title: "Test Page",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_type is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageCreateResponse{
			Page: domain.WikiPage{
				ID:       123,
				Slug:     "test/page",
				Title:    "Test Page",
				PageType: "page",
			},
		}

		mockAdapter.EXPECT().
			CreatePage(gomock.Any(), &domain.WikiPageCreateRequest{
				Slug:     "test/page",
				Title:    "Test Page",
				PageType: "page",
			}).
			Return(expectedResp, nil)

		result, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:     "test/page",
			Title:    "Test Page",
			PageType: "page",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "test/page", result.Slug)
		assert.Equal(t, "Test Page", result.Title)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageCreateResponse{
			Page: domain.WikiPage{
				ID:       456,
				Slug:     "full/page",
				Title:    "Full Page",
				PageType: "page",
				Content:  "Page content",
			},
		}

		mockAdapter.EXPECT().
			CreatePage(gomock.Any(), &domain.WikiPageCreateRequest{
				Slug:       "full/page",
				Title:      "Full Page",
				PageType:   "page",
				Content:    "Page content",
				IsSilent:   true,
				Fields:     []string{"content", "attributes"},
				GridFormat: "yfm",
			}).
			Return(expectedResp, nil)

		result, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:       "full/page",
			Title:      "Full Page",
			PageType:   "page",
			Content:    "Page content",
			IsSilent:   true,
			Fields:     []string{"content", "attributes"},
			GridFormat: "yfm",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(456), result.ID)
	})

	t.Run("adapter/call_with_cloud_page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageCreateResponse{
			Page: domain.WikiPage{
				ID:       789,
				Slug:     "cloud/doc",
				Title:    "Cloud Document",
				PageType: "cloud_page",
			},
		}

		mockAdapter.EXPECT().
			CreatePage(gomock.Any(), &domain.WikiPageCreateRequest{
				Slug:     "cloud/doc",
				Title:    "Cloud Document",
				PageType: "cloud_page",
				CloudPage: &domain.WikiCloudPageInput{
					Method:  "empty_doc",
					Doctype: "docx",
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:     "cloud/doc",
			Title:    "Cloud Document",
			PageType: "cloud_page",
			CloudPage: &CloudPageInput{
				Method:  "empty_doc",
				Doctype: "docx",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(789), result.ID)
		assert.Equal(t, "cloud_page", result.PageType)
	})

	t.Run("result/maps_page_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageCreateResponse{
			Page: domain.WikiPage{
				ID:       100,
				Slug:     "mapped/page",
				Title:    "Mapped Page",
				PageType: "page",
				Content:  "Some content",
				Attributes: &domain.WikiAttributes{
					CommentsCount:   5,
					CommentsEnabled: true,
					CreatedAt:       "2024-01-01T00:00:00Z",
					Lang:            "en",
				},
			},
		}

		mockAdapter.EXPECT().
			CreatePage(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:     "mapped/page",
			Title:    "Mapped Page",
			PageType: "page",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(100), result.ID)
		assert.Equal(t, "mapped/page", result.Slug)
		assert.Equal(t, "Mapped Page", result.Title)
		assert.Equal(t, "page", result.PageType)
		assert.Equal(t, "Some content", result.Content)
		require.NotNil(t, result.Attributes)
		assert.Equal(t, 5, result.Attributes.CommentsCount)
		assert.True(t, result.Attributes.CommentsEnabled)
		assert.Equal(t, "en", result.Attributes.Lang)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"CreatePage",
			400,
			"bad_request",
			"Invalid page type",
			"body with Authorization: Bearer secret",
		)

		mockAdapter.EXPECT().
			CreatePage(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.CreatePage(context.Background(), CreatePageInput{
			Slug:     "test",
			Title:    "Test",
			PageType: "invalid",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 400")
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret")
	})
}

func TestTools_UpdatePage(t *testing.T) {
	t.Parallel()

	t.Run("validation/page_id_zero", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 0,
			Title:  "Updated Title",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("validation/page_id_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: -1,
			Title:  "Updated Title",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("validation/no_fields_to_update", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 123,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one of title, content, or redirect is required")
	})

	t.Run("adapter/call_with_title_only", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageUpdateResponse{
			Page: domain.WikiPage{
				ID:    123,
				Slug:  "test/page",
				Title: "Updated Title",
			},
		}

		mockAdapter.EXPECT().
			UpdatePage(gomock.Any(), &domain.WikiPageUpdateRequest{
				PageID: 123,
				Title:  "Updated Title",
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 123,
			Title:  "Updated Title",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", result.Title)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageUpdateResponse{
			Page: domain.WikiPage{
				ID:      123,
				Slug:    "test/page",
				Title:   "Full Update",
				Content: "Updated content",
			},
		}

		mockAdapter.EXPECT().
			UpdatePage(gomock.Any(), &domain.WikiPageUpdateRequest{
				PageID:     123,
				Title:      "Full Update",
				Content:    "Updated content",
				AllowMerge: true,
				IsSilent:   true,
				Fields:     []string{"content", "attributes"},
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID:     123,
			Title:      "Full Update",
			Content:    "Updated content",
			AllowMerge: true,
			IsSilent:   true,
			Fields:     []string{"content", "attributes"},
		})
		require.NoError(t, err)
		assert.Equal(t, "Full Update", result.Title)
	})

	t.Run("adapter/call_with_redirect", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		targetPageID := int64(456)
		expectedResp := &domain.WikiPageUpdateResponse{
			Page: domain.WikiPage{
				ID:    123,
				Slug:  "old/page",
				Title: "Old Page",
				Redirect: &domain.WikiRedirect{
					PageID: 456,
					Slug:   "new/page",
				},
			},
		}

		mockAdapter.EXPECT().
			UpdatePage(gomock.Any(), &domain.WikiPageUpdateRequest{
				PageID: 123,
				Redirect: &domain.WikiRedirectInput{
					PageID: &targetPageID,
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 123,
			Redirect: &RedirectInput{
				PageID: &targetPageID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result.Redirect)
		assert.Equal(t, int64(456), result.Redirect.PageID)
	})

	t.Run("result/maps_page_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageUpdateResponse{
			Page: domain.WikiPage{
				ID:       123,
				Slug:     "test/page",
				Title:    "Mapped Page",
				PageType: "page",
				Content:  "Content",
				Attributes: &domain.WikiAttributes{
					ModifiedAt: "2024-01-02T00:00:00Z",
					IsReadonly: false,
				},
			},
		}

		mockAdapter.EXPECT().
			UpdatePage(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 123,
			Title:  "Mapped Page",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "test/page", result.Slug)
		assert.Equal(t, "Mapped Page", result.Title)
		require.NotNil(t, result.Attributes)
		assert.Equal(t, "2024-01-02T00:00:00Z", result.Attributes.ModifiedAt)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"UpdatePage",
			404,
			"not_found",
			"Page not found",
			"body",
		)

		mockAdapter.EXPECT().
			UpdatePage(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.UpdatePage(context.Background(), UpdatePageInput{
			PageID: 999,
			Title:  "Updated",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 404")
	})
}

func TestTools_AppendPage(t *testing.T) {
	t.Parallel()

	t.Run("validation/page_id_zero", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  0,
			Content: "Appended content",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("validation/page_id_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  -1,
			Content: "Appended content",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page_id must be positive")
	})

	t.Run("validation/content_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID: 123,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "content is required")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:      123,
				Slug:    "test/page",
				Title:   "Test Page",
				Content: "Original + Appended",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), &domain.WikiPageAppendRequest{
				PageID:  123,
				Content: "Appended content",
			}).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Appended content",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:      123,
				Slug:    "test/page",
				Title:   "Test Page",
				Content: "Updated content",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), &domain.WikiPageAppendRequest{
				PageID:   123,
				Content:  "Full append",
				IsSilent: true,
				Fields:   []string{"content"},
			}).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:   123,
			Content:  "Full append",
			IsSilent: true,
			Fields:   []string{"content"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
	})

	t.Run("adapter/call_with_body_location", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:    123,
				Title: "Test Page",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), &domain.WikiPageAppendRequest{
				PageID:  123,
				Content: "Top content",
				Body: &domain.WikiBodyLocation{
					Location: "top",
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Top content",
			Body: &BodyLocation{
				Location: "top",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
	})

	t.Run("adapter/call_with_section_location", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:    123,
				Title: "Test Page",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), &domain.WikiPageAppendRequest{
				PageID:  123,
				Content: "Section content",
				Section: &domain.WikiSectionLocation{
					ID:       5,
					Location: "bottom",
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Section content",
			Section: &SectionLocation{
				ID:       5,
				Location: "bottom",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
	})

	t.Run("adapter/call_with_anchor_location", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:    123,
				Title: "Test Page",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), &domain.WikiPageAppendRequest{
				PageID:  123,
				Content: "Anchor content",
				Anchor: &domain.WikiAnchorLocation{
					Name:     "my-anchor",
					Fallback: true,
					Regex:    false,
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Anchor content",
			Anchor: &AnchorLocation{
				Name:     "my-anchor",
				Fallback: true,
				Regex:    false,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
	})

	t.Run("result/maps_page_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiPageAppendResponse{
			Page: domain.WikiPage{
				ID:       123,
				Slug:     "test/page",
				Title:    "Test Page",
				PageType: "page",
				Content:  "Full content after append",
			},
		}

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Appended",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "test/page", result.Slug)
		assert.Equal(t, "Test Page", result.Title)
		assert.Equal(t, "page", result.PageType)
		assert.Equal(t, "Full content after append", result.Content)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"AppendPage",
			409,
			"conflict",
			"Page conflict",
			"body",
		)

		mockAdapter.EXPECT().
			AppendPage(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.AppendPage(context.Background(), AppendPageInput{
			PageID:  123,
			Content: "Content",
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 409")
	})
}

func TestTools_CreateGrid(t *testing.T) {
	t.Parallel()

	t.Run("validation/page_id_and_slug_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:    PageInput{},
			Title:   "Test Grid",
			Columns: []ColumnInputCreate{{Slug: "col1", Title: "Column 1"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page.id or page.slug is required")
	})

	t.Run("validation/title_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:    PageInput{ID: 123},
			Columns: []ColumnInputCreate{{Slug: "col1", Title: "Column 1"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("validation/columns_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:    PageInput{ID: 123},
			Title:   "Test Grid",
			Columns: []ColumnInputCreate{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one column is required")
	})

	t.Run("validation/column_slug_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:    PageInput{ID: 123},
			Title:   "Test Grid",
			Columns: []ColumnInputCreate{{Slug: "", Title: "Column 1"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "column slug is required")
	})

	t.Run("validation/column_title_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:    PageInput{ID: 123},
			Title:   "Test Grid",
			Columns: []ColumnInputCreate{{Slug: "col1", Title: ""}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "column title is required")
	})

	t.Run("adapter/call_with_page_id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCreateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-123",
				Title:    "Test Grid",
				Revision: "1",
			},
		}

		mockAdapter.EXPECT().
			CreateGrid(gomock.Any(), &domain.WikiGridCreateRequest{
				PageID: 123,
				Title:  "Test Grid",
				Columns: []domain.WikiColumnDefinition{
					{Slug: "col1", Title: "Column 1", Type: "string"},
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{ID: 123},
			Title: "Test Grid",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1", Type: "string"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-123", result.ID)
		assert.Equal(t, "Test Grid", result.Title)
	})

	t.Run("adapter/call_with_page_slug_resolution", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		resolvedPage := &domain.WikiPage{
			ID:    456,
			Slug:  "target/page",
			Title: "Target Page",
		}

		expectedResp := &domain.WikiGridCreateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-456",
				Title:    "Slug Grid",
				Revision: "1",
			},
		}

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "target/page", domain.WikiGetPageOpts{}).
			Return(resolvedPage, nil)

		mockAdapter.EXPECT().
			CreateGrid(gomock.Any(), &domain.WikiGridCreateRequest{
				PageID: 456,
				Title:  "Slug Grid",
				Columns: []domain.WikiColumnDefinition{
					{Slug: "col1", Title: "Column 1"},
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{Slug: "target/page"},
			Title: "Slug Grid",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-456", result.ID)
	})

	t.Run("adapter/call_with_all_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCreateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-789",
				Title:    "Full Grid",
				Revision: "1",
			},
		}

		mockAdapter.EXPECT().
			CreateGrid(gomock.Any(), &domain.WikiGridCreateRequest{
				PageID: 123,
				Title:  "Full Grid",
				Columns: []domain.WikiColumnDefinition{
					{Slug: "col1", Title: "Column 1", Type: "string"},
					{Slug: "col2", Title: "Column 2", Type: "number"},
				},
				Fields: []string{"attributes", "user_permissions"},
			}).
			Return(expectedResp, nil)

		result, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{ID: 123},
			Title: "Full Grid",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1", Type: "string"},
				{Slug: "col2", Title: "Column 2", Type: "number"},
			},
			Fields: "attributes, user_permissions",
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-789", result.ID)
	})

	t.Run("result/maps_grid_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCreateResponse{
			Grid: domain.WikiGrid{
				ID:             "grid-mapped",
				Title:          "Mapped Grid",
				Revision:       "1",
				RichTextFormat: "yfm",
				CreatedAt:      "2024-01-01T00:00:00Z",
				Structure: []domain.WikiColumn{
					{Slug: "col1", Title: "Column 1", Type: "string"},
				},
				Attributes: &domain.WikiAttributes{
					CreatedAt: "2024-01-01T00:00:00Z",
				},
			},
		}

		mockAdapter.EXPECT().
			CreateGrid(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{ID: 123},
			Title: "Mapped Grid",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1", Type: "string"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-mapped", result.ID)
		assert.Equal(t, "Mapped Grid", result.Title)
		assert.Equal(t, "1", result.Revision)
		assert.Equal(t, "yfm", result.RichTextFmt)
		assert.Equal(t, "2024-01-01T00:00:00Z", result.CreatedAt)
		require.Len(t, result.Structure, 1)
		assert.Equal(t, "col1", result.Structure[0].Slug)
		require.NotNil(t, result.Attributes)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"CreateGrid",
			400,
			"bad_request",
			"Invalid column type",
			"body",
		)

		mockAdapter.EXPECT().
			CreateGrid(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{ID: 123},
			Title: "Test",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1"},
			},
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 400")
	})

	t.Run("error/page_slug_resolution_fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"GetPageBySlug",
			404,
			"not_found",
			"Page not found",
			"body",
		)

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "missing/page", domain.WikiGetPageOpts{}).
			Return(nil, upstreamErr)

		_, err := reg.CreateGrid(context.Background(), CreateGridInput{
			Page:  PageInput{Slug: "missing/page"},
			Title: "Test",
			Columns: []ColumnInputCreate{
				{Slug: "col1", Title: "Column 1"},
			},
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 404")
	})
}

func TestTools_UpdateGridCells(t *testing.T) {
	t.Parallel()

	t.Run("validation/grid_id_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			Cells: []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: "val"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "grid_id is required")
	})

	t.Run("validation/cells_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one cell is required")
	})

	t.Run("validation/cell_row_id_zero", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: 0, ColumnSlug: "col1", Value: "val"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cell[0]: row_id must be positive")
	})

	t.Run("validation/cell_row_id_negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: -1, ColumnSlug: "col1", Value: "val"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cell[0]: row_id must be positive")
	})

	t.Run("validation/cell_column_slug_empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: 1, ColumnSlug: "", Value: "val"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cell[0]: column_slug is required")
	})

	t.Run("validation/cell_value_not_string", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: 123}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cell[0]: value must be a string")
	})

	t.Run("adapter/call_with_minimal_params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCellsUpdateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-123",
				Title:    "Updated Grid",
				Revision: "2",
			},
		}

		mockAdapter.EXPECT().
			UpdateGridCells(gomock.Any(), &domain.WikiGridCellsUpdateRequest{
				GridID: "grid-123",
				Cells: []domain.WikiCellUpdate{
					{RowID: 1, ColumnSlug: "col1", Value: "new value"},
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: "new value"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-123", result.ID)
		assert.Equal(t, "2", result.Revision)
	})

	t.Run("adapter/call_with_multiple_cells", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCellsUpdateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-123",
				Title:    "Multi Updated Grid",
				Revision: "3",
			},
		}

		mockAdapter.EXPECT().
			UpdateGridCells(gomock.Any(), &domain.WikiGridCellsUpdateRequest{
				GridID: "grid-123",
				Cells: []domain.WikiCellUpdate{
					{RowID: 1, ColumnSlug: "col1", Value: "value1"},
					{RowID: 2, ColumnSlug: "col1", Value: "value2"},
					{RowID: 1, ColumnSlug: "col2", Value: "value3"},
				},
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells: []CellUpdateInput{
				{RowID: 1, ColumnSlug: "col1", Value: "value1"},
				{RowID: 2, ColumnSlug: "col1", Value: "value2"},
				{RowID: 1, ColumnSlug: "col2", Value: "value3"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-123", result.ID)
	})

	t.Run("adapter/call_with_revision", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCellsUpdateResponse{
			Grid: domain.WikiGrid{
				ID:       "grid-123",
				Title:    "Grid",
				Revision: "4",
			},
		}

		mockAdapter.EXPECT().
			UpdateGridCells(gomock.Any(), &domain.WikiGridCellsUpdateRequest{
				GridID: "grid-123",
				Cells: []domain.WikiCellUpdate{
					{RowID: 1, ColumnSlug: "col1", Value: "val"},
				},
				Revision: "3",
			}).
			Return(expectedResp, nil)

		result, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID:   "grid-123",
			Cells:    []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: "val"}},
			Revision: "3",
		})
		require.NoError(t, err)
		assert.Equal(t, "4", result.Revision)
	})

	t.Run("result/maps_grid_output", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedResp := &domain.WikiGridCellsUpdateResponse{
			Grid: domain.WikiGrid{
				ID:             "grid-mapped",
				Title:          "Mapped Grid",
				Revision:       "5",
				RichTextFormat: "plain",
				CreatedAt:      "2024-01-01T00:00:00Z",
				Structure: []domain.WikiColumn{
					{Slug: "col1", Title: "Column 1", Type: "string"},
				},
				Rows: []domain.WikiGridRow{
					{ID: "row1", Cells: map[string]domain.WikiGridCell{"col1": {Value: "updated"}}},
				},
			},
		}

		mockAdapter.EXPECT().
			UpdateGridCells(gomock.Any(), gomock.Any()).
			Return(expectedResp, nil)

		result, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-mapped",
			Cells:  []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: "updated"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "grid-mapped", result.ID)
		assert.Equal(t, "Mapped Grid", result.Title)
		assert.Equal(t, "5", result.Revision)
		assert.Equal(t, "plain", result.RichTextFmt)
		require.Len(t, result.Structure, 1)
		require.Len(t, result.Rows, 1)
	})

	t.Run("error/upstream_error_shaped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		upstreamErr := domain.NewUpstreamError(
			domain.ServiceWiki,
			"UpdateGridCells",
			409,
			"conflict",
			"Revision mismatch",
			"body",
		)

		mockAdapter.EXPECT().
			UpdateGridCells(gomock.Any(), gomock.Any()).
			Return(nil, upstreamErr)

		_, err := reg.UpdateGridCells(context.Background(), UpdateGridCellsInput{
			GridID: "grid-123",
			Cells:  []CellUpdateInput{{RowID: 1, ColumnSlug: "col1", Value: "val"}},
		})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "wiki")
		assert.Contains(t, errStr, "HTTP 409")
	})
}
