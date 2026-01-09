//nolint:exhaustruct // test file uses partial struct initialization for clarity
package wiki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestConfig(baseURL, orgID string) *config.Config {
	return &config.Config{ //nolint:exhaustruct // test helper
		WikiBaseURL: baseURL,
		CloudOrgID:  orgID,
	}
}

func TestClient_HeaderInjection(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	const (
		testToken = "test-iam-token"
		testOrgID = "test-org-id"
	)

	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Page{ID: 1, Title: "Test"})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return(testToken, nil)

	client := NewClient(newTestConfig(server.URL, testOrgID), tokenProvider)

	_, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.NoError(t, err)

	assert.Equal(t, "Bearer "+testToken, capturedHeaders.Get(headerAuthorization))
	assert.Equal(t, testOrgID, capturedHeaders.Get(headerCloudOrgID))
	assert.Contains(t, capturedHeaders.Get(headerContentType), "application/json")
}

func TestClient_RetryOnce_ForceRefreshCalled(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	const (
		staleToken = "stale-token"
		freshToken = "fresh-token"
		testOrgID  = "test-org"
	)

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		authHeader := r.Header.Get(headerAuthorization)

		if count == 1 {
			assert.Equal(t, "Bearer "+staleToken, authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error_code":"UNAUTHORIZED","debug_message":"Token expired"}`))
			return
		}

		assert.Equal(t, "Bearer "+freshToken, authHeader)
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Page{ID: 1, Title: "Success"})
	}))
	defer server.Close()

	gomock.InOrder(
		tokenProvider.EXPECT().Token(gomock.Any()).Return(staleToken, nil),
		tokenProvider.EXPECT().ForceRefresh(gomock.Any()).Return(freshToken, nil),
	)

	client := NewClient(newTestConfig(server.URL, testOrgID), tokenProvider)

	page, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), page.ID)
	assert.Equal(t, "Success", page.Title)
	assert.Equal(t, int32(2), requestCount.Load())
}

func TestClient_Non2xx_ReturnsUpstreamError_Sanitized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error_code":"NOT_FOUND","debug_message":"Page not found"}`))
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	_, err := client.GetPageByID(context.Background(), 123, domain.WikiGetPageOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)

	assert.Equal(t, domain.ServiceWiki, upstreamErr.Service)
	assert.Equal(t, "GetPageByID", upstreamErr.Operation)
	assert.Equal(t, http.StatusNotFound, upstreamErr.HTTPStatus)
	assert.Equal(t, "NOT_FOUND", upstreamErr.Code)
	assert.Equal(t, "Page not found", upstreamErr.Message)
	assert.NotContains(t, upstreamErr.Details, "Authorization")
}

func TestClient_Non2xx_FallbackMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal error"))
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	_, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)

	assert.Equal(t, http.StatusInternalServerError, upstreamErr.HTTPStatus)
	assert.Equal(t, "Internal Server Error", upstreamErr.Message)
}

func TestClient_GetPageBySlug_Fields(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Page{ID: 1, Slug: "test/page"})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	page, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{Fields: []string{"content", "attributes"}})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "slug=test%2Fpage")
	assert.Contains(t, capturedURL, "fields=content%2Cattributes")
	assert.Equal(t, "test/page", page.Slug)
}

func TestClient_ListPageResources_Pagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		resp := resourcesResponse{
			Items: []Resource{
				{Type: "attachment", Item: map[string]any{"id": float64(1), "name": "file.txt"}},
			},
			NextCursor: "next-cursor-abc",
			PrevCursor: "prev-cursor-xyz",
		}
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.ListPageResources(context.Background(), 42, domain.WikiListResourcesOpts{
		Cursor:         "start-cursor",
		PageSize:       25,
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Query:          "file",
		Types:          "attachment,grid",
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "/v1/pages/42/resources")
	assert.Contains(t, capturedURL, "cursor=start-cursor")
	assert.Contains(t, capturedURL, "page_size=25")
	assert.Contains(t, capturedURL, "order_by=created_at")
	assert.Contains(t, capturedURL, "order_direction=desc")
	assert.Contains(t, capturedURL, "q=file")
	assert.Contains(t, capturedURL, "types=attachment%2Cgrid")

	assert.Len(t, result.Resources, 1)
	assert.Equal(t, "attachment", result.Resources[0].Type)
	assert.Equal(t, "next-cursor-abc", result.NextCursor)
	assert.Equal(t, "prev-cursor-xyz", result.PrevCursor)
}

func TestClient_ListPageResources_EnforcesMaxPageSize(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resourcesResponse{
			Items:      nil,
			NextCursor: "",
			PrevCursor: "",
		})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	_, err := client.ListPageResources(context.Background(), 1, domain.WikiListResourcesOpts{
		Cursor:         "",
		PageSize:       100, // exceeds max of 50
		OrderBy:        "",
		OrderDirection: "",
		Query:          "",
		Types:          "",
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "page_size=50")
}

func TestClient_ListPageResources_ResourceUnionMapping(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := resourcesResponse{
			Items: []Resource{
				{
					Type: "attachment",
					Item: map[string]any{
						"id":           float64(101),
						"name":         "doc.pdf",
						"size":         float64(2048),
						"mimetype":     "application/pdf",
						"download_url": "https://example.com/doc.pdf",
						"created_at":   "2025-01-01T10:00:00Z",
						"has_preview":  true,
					},
				},
				{
					Type: "sharepoint_resource",
					Item: map[string]any{
						"id":         float64(202),
						"title":      "Shared Doc",
						"doctype":    "spreadsheet",
						"created_at": "2025-02-01T12:00:00Z",
					},
				},
				{
					Type: "grid",
					Item: map[string]any{
						"id":         "grid-uuid-303",
						"title":      "My Grid",
						"created_at": "2025-03-01T14:00:00Z",
					},
				},
			},
			NextCursor: "",
			PrevCursor: "",
		}
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.ListPageResources(context.Background(), 1, domain.WikiListResourcesOpts{
		Cursor:         "",
		PageSize:       0,
		OrderBy:        "",
		OrderDirection: "",
		Query:          "",
		Types:          "",
	})
	require.NoError(t, err)
	require.Len(t, result.Resources, 3)

	// Attachment
	att := result.Resources[0]
	assert.Equal(t, "attachment", att.Type)
	require.NotNil(t, att.Attachment)
	assert.Nil(t, att.Sharepoint)
	assert.Nil(t, att.Grid)
	assert.Equal(t, int64(101), att.Attachment.ID)
	assert.Equal(t, "doc.pdf", att.Attachment.Name)
	assert.Equal(t, int64(2048), att.Attachment.Size)
	assert.Equal(t, "application/pdf", att.Attachment.MIMEType)
	assert.Equal(t, "https://example.com/doc.pdf", att.Attachment.DownloadURL)
	assert.Equal(t, "2025-01-01T10:00:00Z", att.Attachment.CreatedAt)
	assert.True(t, att.Attachment.HasPreview)

	// Sharepoint
	sp := result.Resources[1]
	assert.Equal(t, "sharepoint_resource", sp.Type)
	assert.Nil(t, sp.Attachment)
	require.NotNil(t, sp.Sharepoint)
	assert.Nil(t, sp.Grid)
	assert.Equal(t, int64(202), sp.Sharepoint.ID)
	assert.Equal(t, "Shared Doc", sp.Sharepoint.Title)
	assert.Equal(t, "spreadsheet", sp.Sharepoint.Doctype)
	assert.Equal(t, "2025-02-01T12:00:00Z", sp.Sharepoint.CreatedAt)

	// Grid
	grid := result.Resources[2]
	assert.Equal(t, "grid", grid.Type)
	assert.Nil(t, grid.Attachment)
	assert.Nil(t, grid.Sharepoint)
	require.NotNil(t, grid.Grid)
	assert.Equal(t, "grid-uuid-303", grid.Grid.ID)
	assert.Equal(t, "My Grid", grid.Grid.Title)
	assert.Equal(t, "2025-03-01T14:00:00Z", grid.Grid.CreatedAt)
}

func TestClient_ListPageGrids_Pagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		resp := gridsResponse{
			Items: []PageGridSummary{
				{ID: "grid-uuid-1", Title: "Grid 1", CreatedAt: "2024-01-01T00:00:00Z"},
			},
			NextCursor: "next",
			PrevCursor: "",
		}
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.ListPageGrids(context.Background(), 99, domain.WikiListGridsOpts{
		Cursor:         "",
		PageSize:       30,
		OrderBy:        "title",
		OrderDirection: "asc",
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "/v1/pages/99/grids")
	assert.Contains(t, capturedURL, "page_size=30")
	assert.Contains(t, capturedURL, "order_by=title")
	assert.Contains(t, capturedURL, "order_direction=asc")

	require.Len(t, result.Grids, 1)
	assert.Equal(t, "grid-uuid-1", result.Grids[0].ID)
	assert.Equal(t, "next", result.NextCursor)
}

func TestClient_GetGridByID_WithOptions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Grid{
			ID:    "abc-123",
			Title: "Test Grid",
		})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	grid, err := client.GetGridByID(context.Background(), "abc-123", domain.WikiGetGridOpts{
		Fields:   []string{"attributes", "user_permissions"},
		Filter:   "status=active",
		OnlyCols: "col1,col2",
		OnlyRows: "row1,row2",
		Revision: 5,
		Sort:     "created_at",
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "/v1/grids/abc-123")
	assert.Contains(t, capturedURL, "fields=attributes%2Cuser_permissions")
	assert.Contains(t, capturedURL, "filter=status%3Dactive")
	assert.Contains(t, capturedURL, "only_cols=col1%2Ccol2")
	assert.Contains(t, capturedURL, "only_rows=row1%2Crow2")
	assert.Contains(t, capturedURL, "revision=5")
	assert.Contains(t, capturedURL, "sort=created_at")

	assert.Equal(t, "abc-123", grid.ID)
	assert.Equal(t, "Test Grid", grid.Title)
}

func TestClient_RetryOnce_StillFailsAfterRefresh(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error_code":"UNAUTHORIZED","debug_message":"Still unauthorized"}`))
	}))
	defer server.Close()

	gomock.InOrder(
		tokenProvider.EXPECT().Token(gomock.Any()).Return("token1", nil),
		tokenProvider.EXPECT().ForceRefresh(gomock.Any()).Return("token2", nil),
	)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	_, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusUnauthorized, upstreamErr.HTTPStatus)

	// Verify exactly 2 requests were made (initial + retry)
	assert.Equal(t, int32(2), requestCount.Load())
}

func TestClient_GetPageByID_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Page{
			ID:       42,
			PageType: "page",
			Slug:     "users/test",
			Title:    "Test Page",
			Content:  "Content here",
		})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	page, err := client.GetPageByID(context.Background(), 42, domain.WikiGetPageOpts{Fields: []string{"content"}})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "/v1/pages/42")
	assert.Contains(t, capturedURL, "fields=content")
	assert.Equal(t, int64(42), page.ID)
	assert.Equal(t, "Test Page", page.Title)
	assert.Equal(t, "Content here", page.Content)
}

func TestClient_UpstreamError_NoTokenLeak(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	const secretToken = "super-secret-iam-token-that-must-not-leak"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_code":"BAD_REQUEST","debug_message":"Invalid request"}`))
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return(secretToken, nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	_, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.Error(t, err)

	errStr := err.Error()
	assert.NotContains(t, errStr, secretToken)
	assert.NotContains(t, errStr, "Bearer")
}

func TestClient_FullConfig(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(Page{ID: 1, Title: "Test"})
	}))
	defer server.Close()

	tokenProvider.EXPECT().Token(gomock.Any()).Return("token", nil)

	// Use full config to verify no issues with all fields set
	cfg := &config.Config{
		WikiBaseURL:           server.URL,
		TrackerBaseURL:        "https://api.tracker.yandex.net",
		CloudOrgID:            "org-123",
		IAMTokenRefreshPeriod: 10 * time.Hour,
	}
	client := NewClient(cfg, tokenProvider)

	_, err := client.GetPageBySlug(context.Background(), "test/page", domain.WikiGetPageOpts{})
	require.NoError(t, err)
}
