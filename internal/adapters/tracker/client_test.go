package tracker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestConfig(baseURL, orgID string) *config.Config {
	return &config.Config{ //nolint:exhaustruct // test helper
		TrackerBaseURL: baseURL,
		CloudOrgID:     orgID,
	}
}

func TestClient_HeaderInjection(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	const (
		testToken = "test-iam-token"
		testOrgID = "test-org-id"
	)

	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(issueDTO{ID: "1", Key: "TEST-1"})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return(testToken, nil)

	client := NewClient(newTestConfig(server.URL, testOrgID), tokenProvider)

	//nolint:exhaustruct // test only checks headers
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.NoError(t, err)

	assert.Equal(t, "Bearer "+testToken, capturedHeaders.Get(apihelpers.HeaderAuthorization))
	assert.Equal(t, testOrgID, capturedHeaders.Get(apihelpers.HeaderCloudOrgID))
	assert.Equal(t, "en", capturedHeaders.Get(headerAcceptLanguage))
}

func TestClient_HeaderInjection_POST(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	const (
		testToken = "test-iam-token"
		testOrgID = "test-org-id"
	)

	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode([]issueDTO{})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return(testToken, nil)

	client := NewClient(newTestConfig(server.URL, testOrgID), tokenProvider)

	//nolint:exhaustruct // test only checks headers
	_, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{})
	require.NoError(t, err)

	assert.Equal(t, "Bearer "+testToken, capturedHeaders.Get(apihelpers.HeaderAuthorization))
	assert.Equal(t, testOrgID, capturedHeaders.Get(apihelpers.HeaderCloudOrgID))
	assert.Equal(t, "en", capturedHeaders.Get(headerAcceptLanguage))
	assert.Equal(t, "application/json", capturedHeaders.Get(apihelpers.HeaderContentType))
}

func TestClient_Non2xx_ReturnsUpstreamError_Sanitized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue not found"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks error conversion
	_, err := client.GetIssue(t.Context(), "TEST-999", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)

	assert.Equal(t, domain.ServiceTracker, upstreamErr.Service)
	assert.Equal(t, "GetIssue", upstreamErr.Operation)
	assert.Equal(t, http.StatusNotFound, upstreamErr.HTTPStatus)
	assert.Equal(t, "Issue not found", upstreamErr.Message)
	assert.NotContains(t, upstreamErr.Details, "Authorization")
}

func TestClient_Non2xx_FallbackMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal error"))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks fallback message
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)

	assert.Equal(t, http.StatusInternalServerError, upstreamErr.HTTPStatus)
	assert.Equal(t, "Internal Server Error", upstreamErr.Message)
}

func TestClient_GetIssue_WithExpand(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(issueDTO{ID: "42", Key: "TEST-42", Summary: "Test Issue"})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	issue, err := client.GetIssue(t.Context(), "TEST-42", domain.TrackerGetIssueOpts{Expand: "attachments"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/TEST-42")
	assert.Contains(t, capturedURL, "expand=attachments")
	assert.Equal(t, "TEST-42", issue.Key)
	assert.Equal(t, "Test Issue", issue.Summary)
}

func TestClient_SearchIssues_StandardPagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerXTotalCount, "100")
		w.Header().Set(headerXTotalPages, "5")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode([]issueDTO{
			{ID: "1", Key: "TEST-1"},
			{ID: "2", Key: "TEST-2"},
		})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test uses partial opts
	result, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{
		Filter:  map[string]string{"queue": "TEST"},
		Order:   "+updated",
		PerPage: 20,
		Page:    2,
		Expand:  "transitions",
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/_search")
	assert.Contains(t, capturedURL, "perPage=20")
	assert.Contains(t, capturedURL, "page=2")
	assert.Contains(t, capturedURL, "expand=transitions")

	assert.Equal(t, map[string]any{"queue": "TEST"}, capturedBody["filter"])
	assert.Equal(t, "+updated", capturedBody["order"])

	require.Len(t, result.Issues, 2)
	assert.Equal(t, "TEST-1", result.Issues[0].Key)
	assert.Equal(t, 100, result.TotalCount)
	assert.Equal(t, 5, result.TotalPages)
}

func TestClient_SearchIssues_ScrollPagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerXScrollID, "scroll-id-abc123")
		w.Header().Set(headerXScrollToken, "scroll-token-xyz")
		w.Header().Set(headerLink, `</v3/issues/_search?scrollId=scroll-id-abc123>; rel="next"`)
		w.Header().Set(headerXTotalCount, "50000")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode([]issueDTO{{ID: "1", Key: "TEST-1"}})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test uses scroll pagination opts
	result, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{
		Query:           "Queue: TEST",
		ScrollType:      "sorted",
		PerScroll:       500,
		ScrollTTLMillis: 120000,
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "scrollType=sorted")
	assert.Contains(t, capturedURL, "perScroll=500")
	assert.Contains(t, capturedURL, "scrollTTLMillis=120000")

	assert.Equal(t, "scroll-id-abc123", result.ScrollID)
	assert.Equal(t, "scroll-token-xyz", result.ScrollToken)
	assert.Contains(t, result.NextLink, "scrollId=scroll-id-abc123")
	assert.Equal(t, 50000, result.TotalCount)
}

func TestClient_SearchIssues_ScrollPagination_SubsequentRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerXScrollID, "scroll-id-next")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode([]issueDTO{{ID: "501", Key: "TEST-501"}})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test uses only scrollID
	result, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{
		ScrollID: "scroll-id-abc123",
	})
	require.NoError(t, err)

	assert.Contains(t, capturedURL, "scrollId=scroll-id-abc123")
	assert.Equal(t, "scroll-id-next", result.ScrollID)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "TEST-501", result.Issues[0].Key)
}

func TestClient_CountIssues_WithFilter(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("5221186"))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test uses only filter
	count, err := client.CountIssues(t.Context(), domain.TrackerCountIssuesOpts{
		Filter: map[string]string{"queue": "JUNE", "assignee": "empty()"},
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/_count")
	assert.Equal(t, map[string]any{"queue": "JUNE", "assignee": "empty()"}, capturedBody["filter"])
	assert.Equal(t, 5221186, count)
}

func TestClient_CountIssues_WithQuery(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("42"))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test uses only query
	count, err := client.CountIssues(t.Context(), domain.TrackerCountIssuesOpts{
		Query: "Queue: TEST Assignee: me()",
	})
	require.NoError(t, err)

	assert.Equal(t, "Queue: TEST Assignee: me()", capturedBody["query"])
	assert.Equal(t, 42, count)
}

func TestClient_ListIssueTransitions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode([]transitionDTO{
			{
				ID:      "start_progress",
				Display: "Start Progress",
				To:      &statusDTO{ID: "2", Key: "inProgress", Display: "In Progress"},
			},
			{
				ID:      "resolve",
				Display: "Resolve",
				To:      &statusDTO{ID: "3", Key: "resolved", Display: "Resolved"},
			},
		})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	transitions, err := client.ListIssueTransitions(t.Context(), "TEST-42")
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/TEST-42/transitions")

	require.Len(t, transitions, 2)
	assert.Equal(t, "start_progress", transitions[0].ID)
	assert.Equal(t, "Start Progress", transitions[0].Display)
	assert.Equal(t, "In Progress", transitions[0].To.Display)
}

func TestClient_ListQueues_WithPagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerXTotalCount, "25")
		w.Header().Set(headerXTotalPages, "3")
		//nolint:errcheck // test helper
		w.Write([]byte(`[
			{"id": 1, "key": "DEV", "name": "Development"},
			{"id": 2, "key": "SUPPORT", "name": "Support"}
		]`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.ListQueues(t.Context(), domain.TrackerListQueuesOpts{
		Expand:  "projects,team",
		PerPage: 10,
		Page:    2,
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/queues/")
	assert.Contains(t, capturedURL, "expand=projects%2Cteam")
	assert.Contains(t, capturedURL, "perPage=10")
	assert.Contains(t, capturedURL, "page=2")

	require.Len(t, result.Queues, 2)
	assert.Equal(t, "DEV", result.Queues[0].Key)
	assert.Equal(t, 25, result.TotalCount)
	assert.Equal(t, 3, result.TotalPages)
}

func TestClient_ListIssueComments_WithPagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerLink, `</v3/issues/TEST-1/comments?id=123>; rel="next"`)
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode([]commentDTO{
			{ID: "100", LongID: "long-100", Text: "First comment"},
			{ID: "101", LongID: "long-101", Text: "Second comment"},
		})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.ListIssueComments(t.Context(), "TEST-1", domain.TrackerListCommentsOpts{
		Expand:  "attachments",
		PerPage: 25,
		ID:      "50",
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/TEST-1/comments")
	assert.Contains(t, capturedURL, "expand=attachments")
	assert.Contains(t, capturedURL, "perPage=25")
	assert.Contains(t, capturedURL, "id=50")

	require.Len(t, result.Comments, 2)
	assert.Equal(t, "100", result.Comments[0].ID)
	assert.Equal(t, "First comment", result.Comments[0].Text)
	assert.Contains(t, result.NextLink, "id=123")
}

func TestClient_GetIssueAttachment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	payload := []byte("attachment payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.GetIssueAttachment(t.Context(), "TEST-1", "4159", "attachment.txt")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/TEST-1/attachments/4159/attachment.txt")
	assert.Equal(t, "attachment.txt", result.FileName)
	assert.Equal(t, "application/pdf", result.ContentType)
	assert.Equal(t, payload, result.Data)
}

func TestClient_GetIssueAttachmentPreview(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedURL string
	var capturedMethod string
	payload := []byte{0x1, 0x2, 0x3}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	result, err := client.GetIssueAttachmentPreview(t.Context(), "TEST-1", "4159")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, capturedMethod)
	assert.Contains(t, capturedURL, "/v3/issues/TEST-1/thumbnails/4159")
	assert.Equal(t, "image/png", result.ContentType)
	assert.Equal(t, payload, result.Data)
}

func TestClient_UpstreamError_NoTokenLeak(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	const secretToken = "super-secret-iam-token-that-must-not-leak"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMessages":["Invalid request"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return(secretToken, nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks token leak
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	errStr := err.Error()
	assert.NotContains(t, errStr, secretToken)
	assert.NotContains(t, errStr, "Bearer")
}

func TestClient_ErrorCodes_401(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errorMessages":["Unauthorized"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil).Times(2)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks 401 handling
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusUnauthorized, upstreamErr.HTTPStatus)
	assert.Equal(t, int32(2), requestCount.Load())
}

func TestClient_ErrorCodes_403(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errorMessages":["Access denied"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil).Times(2)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks 403 handling
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusForbidden, upstreamErr.HTTPStatus)
	assert.Equal(t, "Access denied", upstreamErr.Message)
}

func TestClient_ErrorCodes_404(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue not found"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks 404 handling
	_, err := client.GetIssue(t.Context(), "TEST-999", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusNotFound, upstreamErr.HTTPStatus)
	assert.Equal(t, "Issue not found", upstreamErr.Message)
}

func TestClient_ErrorCodes_422(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errorMessages":["Validation failed"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks 422 handling
	_, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{Query: "invalid"})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusUnprocessableEntity, upstreamErr.HTTPStatus)
	assert.Equal(t, "Validation failed", upstreamErr.Message)
}

func TestClient_ErrorCodes_429(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errorMessages":["Rate limit exceeded"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks 429 handling
	_, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Equal(t, http.StatusTooManyRequests, upstreamErr.HTTPStatus)
	assert.Equal(t, "Rate limit exceeded", upstreamErr.Message)
}

func TestClient_IssueID_PathEscaping(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedRawURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRawURL = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,exhaustruct // test helper
		json.NewEncoder(w).Encode(issueDTO{ID: "1", Key: "TEST-1"})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks path escaping
	_, err := client.GetIssue(t.Context(), "TEST/SPECIAL-1", domain.TrackerGetIssueOpts{})
	require.NoError(t, err)

	// RequestURI should contain properly escaped path
	assert.Contains(t, capturedRawURL, "/v3/issues/TEST%2FSPECIAL-1")
}

func TestClient_SearchIssues_QueryLanguage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode([]issueDTO{})
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks query language
	_, err := client.SearchIssues(t.Context(), domain.TrackerSearchIssuesOpts{
		Query: `epic: notEmpty() Queue: TREK "Sort by": Updated DESC`,
	})
	require.NoError(t, err)

	assert.Equal(t, `epic: notEmpty() Queue: TREK "Sort by": Updated DESC`, capturedBody["query"])
}

func TestClient_ErrorResponse_WithErrorsArray(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tokenProvider := apihelpers.NewMockITokenProvider(ctrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":["Error 1","Error 2"]}`))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	tokenProvider.EXPECT().Token(gomock.Any(), gomock.Any()).Return("token", nil)

	client := NewClient(newTestConfig(server.URL, "org"), tokenProvider)

	//nolint:exhaustruct // test checks error array
	_, err := client.GetIssue(t.Context(), "TEST-1", domain.TrackerGetIssueOpts{})
	require.Error(t, err)

	var upstreamErr domain.UpstreamError
	require.ErrorAs(t, err, &upstreamErr)
	assert.Contains(t, upstreamErr.Message, "Error 1")
	assert.Contains(t, upstreamErr.Message, "Error 2")
}
