package wiki

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageToWikiPage(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns nil", func(t *testing.T) {
		t.Parallel()
		result := pageToWikiPage(nil)
		assert.Nil(t, result)
	})

	t.Run("converts all fields including nested structs", func(t *testing.T) {
		t.Parallel()
		dto := &pageDTO{
			ID:       "123",
			PageType: "wiki_page",
			Slug:     "users/docs/readme",
			Title:    "Readme",
			Content:  "# Hello World",
			Attributes: &attributesDTO{
				CommentsCount:   5,
				CommentsEnabled: true,
				CreatedAt:       "2024-01-01T10:00:00Z",
				IsReadonly:      false,
				Lang:            "en",
				ModifiedAt:      "2024-01-02T12:00:00Z",
				IsCollaborative: true,
				IsDraft:         false,
			},
			Redirect: &redirectDTO{
				PageID: "456",
				Slug:   "users/docs/old-readme",
			},
		}

		result := pageToWikiPage(dto)

		require.NotNil(t, result)
		assert.Equal(t, "123", result.ID)
		assert.Equal(t, "wiki_page", result.PageType)
		assert.Equal(t, "users/docs/readme", result.Slug)
		assert.Equal(t, "Readme", result.Title)
		assert.Equal(t, "# Hello World", result.Content)

		require.NotNil(t, result.Attributes)
		assert.Equal(t, 5, result.Attributes.CommentsCount)
		assert.True(t, result.Attributes.CommentsEnabled)
		assert.Equal(t, "2024-01-01T10:00:00Z", result.Attributes.CreatedAt)
		assert.False(t, result.Attributes.IsReadonly)
		assert.Equal(t, "en", result.Attributes.Lang)
		assert.Equal(t, "2024-01-02T12:00:00Z", result.Attributes.ModifiedAt)
		assert.True(t, result.Attributes.IsCollaborative)
		assert.False(t, result.Attributes.IsDraft)

		require.NotNil(t, result.Redirect)
		assert.Equal(t, "456", result.Redirect.PageID)
		assert.Equal(t, "users/docs/old-readme", result.Redirect.Slug)
	})

	t.Run("handles nil attributes and redirect", func(t *testing.T) {
		t.Parallel()
		dto := &pageDTO{
			ID:         "789",
			PageType:   "wiki_page",
			Slug:       "test",
			Title:      "Test",
			Content:    "",
			Attributes: nil,
			Redirect:   nil,
		}

		result := pageToWikiPage(dto)

		require.NotNil(t, result)
		assert.Equal(t, "789", result.ID)
		assert.Nil(t, result.Attributes)
		assert.Nil(t, result.Redirect)
	})
}

func TestResourceToWikiResource_Polymorphism(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns nil", func(t *testing.T) {
		t.Parallel()
		result, err := resourceToWikiResource(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("attachment type populates Attachment field", func(t *testing.T) {
		t.Parallel()
		// Simulates the map[string]any that json.Unmarshal produces
		dto := &resourceDTO{
			Type: "attachment",
			Item: map[string]any{
				"id":           float64(100),
				"name":         "document.pdf",
				"size":         float64(1024),
				"mimetype":     "application/pdf",
				"download_url": "https://wiki.example.com/download/100",
				"created_at":   "2024-03-15T09:30:00Z",
				"has_preview":  true,
			},
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "attachment", result.Type)

		require.NotNil(t, result.Attachment, "Attachment pointer should be set")
		assert.Equal(t, "100", result.Attachment.ID)
		assert.Equal(t, "document.pdf", result.Attachment.Name)
		assert.Equal(t, int64(1024), result.Attachment.Size)
		assert.Equal(t, "application/pdf", result.Attachment.MIMEType)
		assert.Equal(t, "https://wiki.example.com/download/100", result.Attachment.DownloadURL)
		assert.Equal(t, "2024-03-15T09:30:00Z", result.Attachment.CreatedAt)
		assert.True(t, result.Attachment.HasPreview)

		assert.Nil(t, result.Sharepoint, "Sharepoint pointer should be nil for attachment type")
		assert.Nil(t, result.Grid, "Grid pointer should be nil for attachment type")
	})

	t.Run("sharepoint_resource type populates Sharepoint field", func(t *testing.T) {
		t.Parallel()
		dto := &resourceDTO{
			Type: "sharepoint_resource",
			Item: map[string]any{
				"id":         float64(200),
				"title":      "Budget 2024",
				"doctype":    "xlsx",
				"created_at": "2024-02-20T14:00:00Z",
			},
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "sharepoint_resource", result.Type)

		assert.Nil(t, result.Attachment, "Attachment pointer should be nil for sharepoint_resource type")
		assert.Nil(t, result.Grid, "Grid pointer should be nil for sharepoint_resource type")

		require.NotNil(t, result.Sharepoint, "Sharepoint pointer should be set")
		assert.Equal(t, "200", result.Sharepoint.ID)
		assert.Equal(t, "Budget 2024", result.Sharepoint.Title)
		assert.Equal(t, "xlsx", result.Sharepoint.Doctype)
		assert.Equal(t, "2024-02-20T14:00:00Z", result.Sharepoint.CreatedAt)
	})

	t.Run("grid type populates Grid field", func(t *testing.T) {
		t.Parallel()
		dto := &resourceDTO{
			Type: "grid",
			Item: map[string]any{
				"id":         "550e8400-e29b-41d4-a716-446655440000",
				"title":      "Project Tracker",
				"created_at": "2024-05-10T12:00:00Z",
			},
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "grid", result.Type)

		assert.Nil(t, result.Attachment, "Attachment pointer should be nil for grid type")
		assert.Nil(t, result.Sharepoint, "Sharepoint pointer should be nil for grid type")

		require.NotNil(t, result.Grid, "Grid pointer should be set")
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.Grid.ID)
		assert.Equal(t, "Project Tracker", result.Grid.Title)
		assert.Equal(t, "2024-05-10T12:00:00Z", result.Grid.CreatedAt)
	})

	t.Run("unknown type leaves all pointers nil", func(t *testing.T) {
		t.Parallel()
		dto := &resourceDTO{
			Type: "future_resource_type",
			Item: map[string]any{"some_field": "some_value"},
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "future_resource_type", result.Type)
		assert.Nil(t, result.Attachment)
		assert.Nil(t, result.Sharepoint)
		assert.Nil(t, result.Grid)
	})

	t.Run("nil item leaves all pointers nil", func(t *testing.T) {
		t.Parallel()
		dto := &resourceDTO{
			Type: "attachment",
			Item: nil,
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "attachment", result.Type)
		assert.Nil(t, result.Attachment)
		assert.Nil(t, result.Sharepoint)
		assert.Nil(t, result.Grid)
	})
}

func TestResourcesPageToWikiResourcesPage(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns nil", func(t *testing.T) {
		t.Parallel()
		result, err := resourcesPageToWikiResourcesPage(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("converts multiple resources with pagination cursors", func(t *testing.T) {
		t.Parallel()
		dto := &resourcesPageDTO{
			Resources: []resourceDTO{
				{
					Type: "attachment",
					Item: map[string]any{
						"id":           float64(1),
						"name":         "file1.txt",
						"size":         float64(512),
						"mimetype":     "text/plain",
						"download_url": "https://wiki.example.com/download/1",
						"created_at":   "2024-01-01T00:00:00Z",
						"has_preview":  false,
					},
				},
				{
					Type: "sharepoint_resource",
					Item: map[string]any{
						"id":         float64(2),
						"title":      "Report",
						"doctype":    "docx",
						"created_at": "2024-01-02T00:00:00Z",
					},
				},
			},
			NextCursor: "cursor-next",
			PrevCursor: "cursor-prev",
		}

		result, err := resourcesPageToWikiResourcesPage(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "cursor-next", result.NextCursor)
		assert.Equal(t, "cursor-prev", result.PrevCursor)
		require.Len(t, result.Resources, 2)

		assert.Equal(t, "attachment", result.Resources[0].Type)
		require.NotNil(t, result.Resources[0].Attachment)
		assert.Equal(t, "file1.txt", result.Resources[0].Attachment.Name)

		assert.Equal(t, "sharepoint_resource", result.Resources[1].Type)
		require.NotNil(t, result.Resources[1].Sharepoint)
		assert.Equal(t, "Report", result.Resources[1].Sharepoint.Title)
	})

	t.Run("empty resources slice converts to empty slice", func(t *testing.T) {
		t.Parallel()
		dto := &resourcesPageDTO{
			Resources:  []resourceDTO{},
			NextCursor: "",
			PrevCursor: "",
		}

		result, err := resourcesPageToWikiResourcesPage(dto)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.Resources)
	})
}

func TestGridRowToWikiGridRow_CellsMapping(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()
		result := gridRowToWikiGridRow(nil)
		assert.Empty(t, result.ID)
		assert.Nil(t, result.Cells)
	})

	t.Run("converts various cell value types to strings", func(t *testing.T) {
		t.Parallel()
		dto := &gridRowDTO{
			ID: "row-123",
			Cells: map[string]any{
				"text_col":    "hello world",
				"number_col":  float64(42),
				"float_col":   float64(3.14),
				"bool_col":    true,
				"null_col":    nil,
				"complex_col": map[string]any{"nested": "value"},
			},
		}

		result := gridRowToWikiGridRow(dto)

		assert.Equal(t, "row-123", result.ID)
		require.Len(t, result.Cells, 6)

		assert.Equal(t, "hello world", result.Cells["text_col"].Value)
		assert.Equal(t, "42", result.Cells["number_col"].Value)
		assert.Equal(t, "3.14", result.Cells["float_col"].Value)
		assert.Equal(t, "true", result.Cells["bool_col"].Value)
		assert.Empty(t, result.Cells["null_col"].Value)
		assert.JSONEq(t, `{"nested":"value"}`, result.Cells["complex_col"].Value)
	})

	t.Run("empty cells map converts to empty map", func(t *testing.T) {
		t.Parallel()
		dto := &gridRowDTO{
			ID:    "empty-row",
			Cells: map[string]any{},
		}

		result := gridRowToWikiGridRow(dto)

		assert.Equal(t, "empty-row", result.ID)
		assert.Empty(t, result.Cells)
	})
}

func TestGridToWikiGrid(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns nil", func(t *testing.T) {
		t.Parallel()
		result := gridToWikiGrid(nil)
		assert.Nil(t, result)
	})

	t.Run("converts grid with structure and rows", func(t *testing.T) {
		t.Parallel()
		dto := &gridDTO{
			ID:    "grid-456",
			Title: "Employee Directory",
			Structure: []columnDTO{
				{Slug: "name", Title: "Name", Type: "string"},
				{Slug: "age", Title: "Age", Type: "number"},
			},
			Rows: []gridRowDTO{
				{
					ID: "row-1",
					Cells: map[string]any{
						"name": "Alice",
						"age":  float64(30),
					},
				},
				{
					ID: "row-2",
					Cells: map[string]any{
						"name": "Bob",
						"age":  float64(25),
					},
				},
			},
			Revision:       "rev-789",
			CreatedAt:      "2024-05-01T08:00:00Z",
			RichTextFormat: "markdown",
			Attributes: &attributesDTO{
				CommentsCount:   10,
				CommentsEnabled: true,
				CreatedAt:       "2024-05-01T08:00:00Z",
				IsReadonly:      false,
				Lang:            "en",
				ModifiedAt:      "2024-05-10T16:00:00Z",
				IsCollaborative: true,
				IsDraft:         false,
			},
		}

		result := gridToWikiGrid(dto)

		require.NotNil(t, result)
		assert.Equal(t, "grid-456", result.ID)
		assert.Equal(t, "Employee Directory", result.Title)
		assert.Equal(t, "rev-789", result.Revision)
		assert.Equal(t, "2024-05-01T08:00:00Z", result.CreatedAt)
		assert.Equal(t, "markdown", result.RichTextFormat)

		require.Len(t, result.Structure, 2)
		assert.Equal(t, "name", result.Structure[0].Slug)
		assert.Equal(t, "Name", result.Structure[0].Title)
		assert.Equal(t, "string", result.Structure[0].Type)
		assert.Equal(t, "age", result.Structure[1].Slug)

		require.Len(t, result.Rows, 2)
		assert.Equal(t, "row-1", result.Rows[0].ID)
		assert.Equal(t, "Alice", result.Rows[0].Cells["name"].Value)
		assert.Equal(t, "30", result.Rows[0].Cells["age"].Value)
		assert.Equal(t, "row-2", result.Rows[1].ID)
		assert.Equal(t, "Bob", result.Rows[1].Cells["name"].Value)
		assert.Equal(t, "25", result.Rows[1].Cells["age"].Value)

		require.NotNil(t, result.Attributes)
		assert.Equal(t, 10, result.Attributes.CommentsCount)
	})

	t.Run("handles nil attributes", func(t *testing.T) {
		t.Parallel()
		dto := &gridDTO{
			ID:             "grid-minimal",
			Title:          "Minimal Grid",
			Structure:      []columnDTO{},
			Rows:           []gridRowDTO{},
			Revision:       "1",
			CreatedAt:      "2024-06-01T00:00:00Z",
			RichTextFormat: "",
			Attributes:     nil,
		}

		result := gridToWikiGrid(dto)

		require.NotNil(t, result)
		assert.Equal(t, "grid-minimal", result.ID)
		assert.Empty(t, result.Structure)
		assert.Empty(t, result.Rows)
		assert.Nil(t, result.Attributes)
	})
}

func TestGridsPageToWikiGridsPage(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns nil", func(t *testing.T) {
		t.Parallel()
		result := gridsPageToWikiGridsPage(nil)
		assert.Nil(t, result)
	})

	t.Run("converts grids page with summaries and cursors", func(t *testing.T) {
		t.Parallel()
		dto := &gridsPageDTO{
			Grids: []pageGridSummaryDTO{
				{ID: "grid-1", Title: "Grid One", CreatedAt: "2024-01-01T00:00:00Z"},
				{ID: "grid-2", Title: "Grid Two", CreatedAt: "2024-02-01T00:00:00Z"},
			},
			NextCursor: "next",
			PrevCursor: "prev",
		}

		result := gridsPageToWikiGridsPage(dto)

		require.NotNil(t, result)
		assert.Equal(t, "next", result.NextCursor)
		assert.Equal(t, "prev", result.PrevCursor)
		require.Len(t, result.Grids, 2)
		assert.Equal(t, "grid-1", result.Grids[0].ID)
		assert.Equal(t, "Grid One", result.Grids[0].Title)
		assert.Equal(t, "2024-01-01T00:00:00Z", result.Grids[0].CreatedAt)
		assert.Equal(t, "grid-2", result.Grids[1].ID)
	})
}

func TestCellValueToString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil value", nil, ""},
		{"string value", "text", "text"},
		{"integer as float64", float64(100), "100"},
		{"negative integer as float64", float64(-50), "-50"},
		{"decimal float64", float64(3.14159), "3.14159"},
		{"boolean true", true, "true"},
		{"boolean false", false, "false"},
		{"slice", []any{"a", "b"}, `["a","b"]`},
		{"nested map", map[string]any{"key": "value"}, `{"key":"value"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cellValueToString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAttachmentFieldMapping(t *testing.T) {
	t.Parallel()

	t.Run("Mimetype DTO field maps to MIMEType domain field", func(t *testing.T) {
		t.Parallel()
		dto := &resourceDTO{
			Type: "attachment",
			Item: map[string]any{
				"id":           float64(1),
				"name":         "test.pdf",
				"size":         float64(2048),
				"mimetype":     "application/pdf",
				"download_url": "https://example.com/test.pdf",
				"created_at":   "2024-01-01T00:00:00Z",
				"has_preview":  true,
			},
		}

		result, err := resourceToWikiResource(dto)

		require.NoError(t, err)
		require.NotNil(t, result.Attachment)
		// DTO field is "mimetype", domain field is "MIMEType"
		assert.Equal(t, "application/pdf", result.Attachment.MIMEType)
	})
}

// TestConversionsPreserveAllFields validates that no fields are lost in conversion.
func TestConversionsPreserveAllFields(t *testing.T) {
	t.Parallel()

	t.Run("WikiAttributes preserves all 8 fields", func(t *testing.T) {
		t.Parallel()
		dto := &attributesDTO{
			CommentsCount:   42,
			CommentsEnabled: true,
			CreatedAt:       "created",
			IsReadonly:      true,
			Lang:            "ru",
			ModifiedAt:      "modified",
			IsCollaborative: true,
			IsDraft:         true,
		}

		result := attributesToWikiAttributes(dto)

		require.NotNil(t, result)
		assert.Equal(t, 42, result.CommentsCount)
		assert.True(t, result.CommentsEnabled)
		assert.Equal(t, "created", result.CreatedAt)
		assert.True(t, result.IsReadonly)
		assert.Equal(t, "ru", result.Lang)
		assert.Equal(t, "modified", result.ModifiedAt)
		assert.True(t, result.IsCollaborative)
		assert.True(t, result.IsDraft)
	})

	t.Run("WikiAttachment preserves all 7 fields", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"id":           float64(999),
			"name":         "attachment.zip",
			"size":         float64(999999),
			"mimetype":     "application/zip",
			"download_url": "https://example.com/attachment.zip",
			"created_at":   "2024-12-31T23:59:59Z",
			"has_preview":  false,
		}

		result, err := convertItemToAttachment(item)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "999", result.ID)
		assert.Equal(t, "attachment.zip", result.Name)
		assert.Equal(t, int64(999999), result.Size)
		assert.Equal(t, "application/zip", result.MIMEType)
		assert.Equal(t, "https://example.com/attachment.zip", result.DownloadURL)
		assert.Equal(t, "2024-12-31T23:59:59Z", result.CreatedAt)
		assert.False(t, result.HasPreview)
	})

	t.Run("WikiSharepointResource preserves all 4 fields", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"id":         float64(777),
			"title":      "SharePoint Doc",
			"doctype":    "pptx",
			"created_at": "2024-06-15T12:00:00Z",
		}

		result, err := convertItemToSharepoint(item)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "777", result.ID)
		assert.Equal(t, "SharePoint Doc", result.Title)
		assert.Equal(t, "pptx", result.Doctype)
		assert.Equal(t, "2024-06-15T12:00:00Z", result.CreatedAt)
	})

	t.Run("WikiColumn preserves all 3 fields", func(t *testing.T) {
		t.Parallel()
		dto := &columnDTO{
			Slug:  "col_slug",
			Title: "Column Title",
			Type:  "date",
		}

		result := columnToWikiColumn(dto)

		assert.Equal(t, "col_slug", result.Slug)
		assert.Equal(t, "Column Title", result.Title)
		assert.Equal(t, "date", result.Type)
	})
}
