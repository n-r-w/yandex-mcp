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

		_, err := reg.getPageBySlug(context.Background(), getPageBySlugInputDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slug is required")
	})

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:       "123",
			PageType: "doc",
			Slug:     "test/page",
			Title:    "Test Page",
			Content:  "Hello",
		}

		mockAdapter.EXPECT().
			GetPageBySlug(gomock.Any(), "test/page", domain.WikiGetPageOpts{Fields: []string{"content", "attributes"}}).
			Return(expectedPage, nil)

		input := getPageBySlugInputDTO{
			Slug:   "test/page",
			Fields: []string{"content", "attributes"},
		}

		result, err := reg.getPageBySlug(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, "123", result.ID)
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

		_, err := reg.getPageBySlug(context.Background(), getPageBySlugInputDTO{Slug: "missing"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), domain.ServiceWiki)
		assert.Contains(t, err.Error(), "GetPageBySlug")
		assert.Contains(t, err.Error(), "HTTP 404")
		assert.NotContains(t, err.Error(), "token")
	})
}

func TestTools_GetPageByID(t *testing.T) {
	t.Parallel()

	t.Run("calls adapter with correct parameters", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:       "456",
			PageType: "doc",
			Slug:     "another/page",
			Title:    "Another Page",
		}

		mockAdapter.EXPECT().
			GetPageByID(gomock.Any(), "456", domain.WikiGetPageOpts{Fields: []string{"title"}}).
			Return(expectedPage, nil)

		result, err := reg.getPageByID(context.Background(), getPageByIDInputDTO{
			PageID: "456",
			Fields: []string{"title"},
		})
		require.NoError(t, err)
		assert.Equal(t, "456", result.ID)
		assert.Equal(t, "Another Page", result.Title)
	})

	t.Run("maps attributes correctly", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		expectedPage := &domain.WikiPage{
			ID:    "789",
			Title: "With Attrs",
			Attributes: &domain.WikiAttributes{
				CommentsCount:   5,
				CommentsEnabled: true,
				CreatedAt:       "2024-01-01T00:00:00Z",
				Lang:            "en",
			},
		}

		mockAdapter.EXPECT().
			GetPageByID(gomock.Any(), "789", domain.WikiGetPageOpts{}).
			Return(expectedPage, nil)

		result, err := reg.getPageByID(context.Background(), getPageByIDInputDTO{PageID: "789"})
		require.NoError(t, err)
		require.NotNil(t, result.Attributes)
		assert.Equal(t, 5, result.Attributes.CommentsCount)
		assert.True(t, result.Attributes.CommentsEnabled)
		assert.Equal(t, "en", result.Attributes.Lang)
	})
}

func TestTools_ListResources(t *testing.T) {
	t.Parallel()

	t.Run("returns error when page_size is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.listResources(context.Background(), listResourcesInputDTO{
			PageID:   "123",
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

		_, err := reg.listResources(context.Background(), listResourcesInputDTO{
			PageID:   "123",
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
			ListPageResources(gomock.Any(), "100", domain.WikiListResourcesOpts{
				Cursor:         "cursor1",
				PageSize:       20,
				OrderBy:        "created_at",
				OrderDirection: "desc",
				Query:          "test",
				Types:          "attachment",
			}).
			Return(expectedResult, nil)

		input := listResourcesInputDTO{
			PageID:         "100",
			Cursor:         "cursor1",
			PageSize:       20,
			OrderBy:        "created_at",
			OrderDirection: "desc",
			Q:              "test",
			Types:          "attachment",
		}

		result, err := reg.listResources(context.Background(), input)
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
						ID:          "101",
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
			ListPageResources(gomock.Any(), "100", gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.listResources(context.Background(), listResourcesInputDTO{PageID: "100"})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "attachment", res.Type)
		require.NotNil(t, res.Item)

		attachment, ok := res.Item.(attachmentOutputDTO)
		require.True(t, ok, "expected AttachmentOutput, got %T", res.Item)
		assert.Equal(t, "101", attachment.ID)
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
						ID:        "202",
						Title:     "Important Document",
						Doctype:   "docx",
						CreatedAt: "2024-02-01T10:00:00Z",
					},
				},
			},
		}

		mockAdapter.EXPECT().
			ListPageResources(gomock.Any(), "100", gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.listResources(context.Background(), listResourcesInputDTO{PageID: "100"})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "sharepoint_resource", res.Type)
		require.NotNil(t, res.Item)

		sharepoint, ok := res.Item.(sharepointResourceOutputDTO)
		require.True(t, ok, "expected SharepointResourceOutput, got %T", res.Item)
		assert.Equal(t, "202", sharepoint.ID)
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
			ListPageResources(gomock.Any(), "100", gomock.Any()).
			Return(expectedResult, nil)

		result, err := reg.listResources(context.Background(), listResourcesInputDTO{PageID: "100"})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)

		res := result.Resources[0]
		assert.Equal(t, "grid", res.Type)
		require.NotNil(t, res.Item)

		grid, ok := res.Item.(gridResourceOutputDTO)
		require.True(t, ok, "expected GridResourceOutput, got %T", res.Item)
		assert.Equal(t, "grid-xyz-123", grid.ID)
		assert.Equal(t, "Sales Data", grid.Title)
		assert.Equal(t, "2024-03-01T15:30:00Z", grid.CreatedAt)
	})
}

func TestTools_ListGrids(t *testing.T) {
	t.Parallel()

	t.Run("returns error when page_size is negative", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockAdapter := NewMockIWikiAdapter(ctrl)
		reg := NewRegistrator(mockAdapter, domain.WikiAllTools())

		_, err := reg.listGrids(context.Background(), listGridsInputDTO{
			PageID:   "123",
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

		_, err := reg.listGrids(context.Background(), listGridsInputDTO{
			PageID:   "123",
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
			ListPageGrids(gomock.Any(), "200", domain.WikiListGridsOpts{
				Cursor:   "cur",
				PageSize: 10,
			}).
			Return(expectedResult, nil)

		result, err := reg.listGrids(context.Background(), listGridsInputDTO{
			PageID:   "200",
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

		_, err := reg.getGrid(context.Background(), getGridInputDTO{})
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
				Revision: "5",
			}).
			Return(expectedGrid, nil)

		result, err := reg.getGrid(context.Background(), getGridInputDTO{
			GridID:   "grid123",
			Fields:   []string{"rows"},
			Filter:   "col1 = 'test'",
			OnlyCols: "col1",
			Revision: "5",
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

		result, err := reg.getGrid(context.Background(), getGridInputDTO{GridID: "gridWithAttrs"})
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

		result, err := reg.getGrid(context.Background(), getGridInputDTO{GridID: "gridCellTypes"})
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

		_, err := reg.getPageBySlug(context.Background(), getPageBySlugInputDTO{Slug: "test"})
		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, domain.ServiceWiki)
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

		_, err := reg.getPageBySlug(context.Background(), getPageBySlugInputDTO{Slug: "test"})
		require.Error(t, err)
		errStr := err.Error()
		// Non-upstream errors should return a generic safe message
		assert.Equal(t, "wiki: internal error", errStr)
		assert.NotContains(t, errStr, "Bearer")
		assert.NotContains(t, errStr, "secret-token-123")
	})
}
