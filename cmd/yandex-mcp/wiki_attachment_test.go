package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Wiki attachment fixture values identify the tool call, path, limits, and credentials.
const (
	wikiAttachmentTool = "wiki_attachment_get"
	wikiAttachmentPath = "Team/ProJect/.files/Q1 %2f%23+résumé.DATA"
	wikiAttachmentCap  = 8
	wikiFixtureToken   = "fixture-initial-token"
	wikiRefreshedToken = "fixture-refreshed-token"
	wikiFixtureOrg     = "fixture-org"
)

// connectWikiAttachmentClient opens an MCP session with a local Wiki API fixture.
func connectWikiAttachmentClient(
	t *testing.T, baseURL, allowedDir string, capBytes int64, token apihelpers.ITokenProvider,
) *mcp.ClientSession {
	t.Helper()
	cfg := &config.Config{
		WikiBaseURL:             baseURL,
		TrackerBaseURL:          baseURL,
		CloudOrgID:              "",
		OrgID:                   wikiFixtureOrg,
		OAuthTokenFile:          "",
		OAuthProfile:            "",
		TokenType:               config.TokenTypeOAuth,
		CLIProfile:              "",
		IAMTokenRefreshPeriod:   0,
		ToolTimeout:             5 * time.Second,
		TokenSource:             "",
		AuthAgentPort:           0,
		AttachAllowedExtensions: []string{"payload"},
		AttachViewExtensions:    []string{"data"},
		AttachAllowedDirs:       []string{allowedDir},
		AttachInlineMaxBytes:    capBytes,
	}
	srv, err := newMCPServer("test", cfg, token)
	require.NoError(t, err)
	client := mcp.NewClient(&mcp.Implementation{ //nolint:exhaustruct_v5 // optional metadata uses defaults
		Name:    "test-client",
		Version: "test",
	}, nil)
	srvTransport, clientTransport := mcp.NewInMemoryTransports()
	_, err = srv.Connect(t.Context(), srvTransport)
	require.NoError(t, err)
	session, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

// callWikiAttachment returns the tool result even when the tool reports a delivery error.
func callWikiAttachment(t *testing.T, session *mcp.ClientSession, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{ //nolint:exhaustruct_v5 // metadata is optional
		Name:      wikiAttachmentTool,
		Arguments: args,
	})
	require.NoError(t, err)
	return result
}

// assertWikiDownloadRequest checks the request contract at the Wiki API boundary.
func assertWikiDownloadRequest(t *testing.T, r *http.Request, token string) {
	t.Helper()
	assert.Equal(t, http.MethodGet, r.Method)
	assert.Equal(t, "/v1/pages/attachments/download_by_url", r.URL.Path)
	assert.Equal(t, url.Values{"url": {wikiAttachmentPath}, "download": {"true"}}, r.URL.Query())
	assert.Equal(t, "OAuth "+token, r.Header.Get("Authorization"))
	assert.Equal(t, wikiFixtureOrg, r.Header.Get("X-Org-Id"))
}

// assertWikiAttachmentError checks that a delivery failure exposes neither contents nor credentials.
func assertWikiAttachmentError(t *testing.T, result *mcp.CallToolResult, message string) {
	t.Helper()
	require.True(t, result.IsError)
	assert.Nil(t, result.StructuredContent)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, content.Text, message)
	for _, secret := range []string{"fixture-debug", wikiFixtureToken, wikiRefreshedToken, wikiFixtureOrg} {
		assert.NotContains(t, content.Text, secret)
	}
}

// TestWikiAttachmentDelivery checks complete inline and disk results under a non-default attachment policy.
func TestWikiAttachmentDelivery(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, errorText string
		capBytes              int64
		save                  bool
	}{
		{name: "inline at limit", body: "éééé", capBytes: wikiAttachmentCap, errorText: "", save: false},
		{name: "empty inline", body: "", capBytes: wikiAttachmentCap, errorText: "", save: false},
		{name: "maximum configured cap", body: "é", capBytes: math.MaxInt64, errorText: "", save: false},
		{
			name: "one byte over limit", body: "éééé!", capBytes: wikiAttachmentCap,
			errorText: "inline limit", save: false,
		},
		{name: "invalid UTF-8", body: "bad\xff", capBytes: wikiAttachmentCap, errorText: "UTF-8", save: false},
		{
			name: "save over limit with binary bytes", body: strings.Repeat("é\x00\xff", 2048),
			capBytes: wikiAttachmentCap, errorText: "", save: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var hits atomic.Int32
			fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				assertWikiDownloadRequest(t, r, wikiFixtureToken)
				_, _ = io.WriteString(w, tc.body)
			}))
			t.Cleanup(fixture.Close)
			token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
			token.EXPECT().Token(gomock.Any(), false).Return(wikiFixtureToken, nil)
			dir := t.TempDir()
			session := connectWikiAttachmentClient(t, fixture.URL, dir, tc.capBytes, token)
			args := map[string]any{"attachment_path": "/" + wikiAttachmentPath}
			path := filepath.Join(dir, "nested", "result.payload")
			if tc.save {
				args["attachment_path"] = wikiAttachmentPath
				args["save_path"] = path
			} else {
				args["get_content"] = true
			}
			result := callWikiAttachment(t, session, args)
			assert.EqualValues(t, 1, hits.Load())
			if tc.errorText != "" {
				assertWikiAttachmentError(t, result, tc.errorText)
				return
			}
			require.False(t, result.IsError, "%+v", result.Content)
			want := map[string]any{"size": float64(len(tc.body)), "content": tc.body}
			if tc.save {
				delete(want, "content")
				want["saved_path"] = path
				data, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, tc.body, string(data))
			}
			assert.Equal(t, want, result.StructuredContent)
		})
	}
}

// TestWikiAttachmentFinalStatus verifies that unsuccessful statuses return no attachment in either mode.
func TestWikiAttachmentFinalStatus(t *testing.T) {
	t.Parallel()
	for _, status := range []int{
		http.StatusUnauthorized, http.StatusForbidden, http.StatusFound, http.StatusPartialContent,
	} {
		for _, save := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/save=%t", status, save), func(t *testing.T) {
				t.Parallel()
				accessDenied := status == http.StatusUnauthorized || status == http.StatusForbidden
				var hits, redirectHits atomic.Int32
				fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/redirect-target" {
						redirectHits.Add(1)
						_, _ = io.WriteString(w, "partial")
						return
					}
					token := wikiFixtureToken
					if hits.Add(1) == 2 {
						token = wikiRefreshedToken
					}
					assertWikiDownloadRequest(t, r, token)
					if status == http.StatusFound {
						w.Header().Set("Location", "/redirect-target")
					}
					w.WriteHeader(status)
					_, _ = io.WriteString(w, strings.Repeat("fixture-debug ", 8192))
				}))
				t.Cleanup(fixture.Close)
				token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
				initial := token.EXPECT().Token(gomock.Any(), false).Return(wikiFixtureToken, nil)
				wantHits, wantError := 1, fmt.Sprintf("HTTP %d", status)
				if accessDenied {
					token.EXPECT().Token(gomock.Any(), true).Return(wikiRefreshedToken, nil).After(initial)
					wantHits, wantError = 2, "access denied"
				}
				dir := t.TempDir()
				path := filepath.Join(dir, "nested", "result.payload")
				session := connectWikiAttachmentClient(t, fixture.URL, dir, wikiAttachmentCap, token)
				args := map[string]any{"attachment_path": wikiAttachmentPath, "get_content": true}
				if save {
					delete(args, "get_content")
					args["save_path"] = path
				}
				assertWikiAttachmentError(t, callWikiAttachment(t, session, args), wantError)
				assert.EqualValues(t, wantHits, hits.Load())
				assert.Zero(t, redirectHits.Load())
				assert.NoDirExists(t, filepath.Dir(path))
			})
		}
	}
}

// TestWikiAttachmentRejectsInputBeforeDownload checks invalid inputs without contacting Wiki.
func TestWikiAttachmentRejectsInputBeforeDownload(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(fixture.Close)
	dir, outside := t.TempDir(), t.TempDir()
	path := filepath.Join(dir, "result.payload")
	existing := filepath.Join(dir, "existing.payload")
	require.NoError(t, os.WriteFile(existing, []byte("keep"), 0o600))
	token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
	// Keep a permissive token mock so a broken preflight reaches the fixture and is counted.
	token.EXPECT().Token(gomock.Any(), gomock.Any()).Return(wikiFixtureToken, nil).AnyTimes()
	session := connectWikiAttachmentClient(t, fixture.URL, dir, wikiAttachmentCap, token)

	inline := func(path string) map[string]any {
		return map[string]any{"attachment_path": path, "get_content": true}
	}
	save := func(destination string) map[string]any {
		return map[string]any{"attachment_path": wikiAttachmentPath, "save_path": destination}
	}
	both := inline(wikiAttachmentPath)
	both["save_path"] = path
	for _, tc := range []struct {
		name, message string
		args          map[string]any
	}{
		{"missing path", "attachment_path", map[string]any{"get_content": true}},
		{"absolute URL", "attachment_path", inline(fixture.URL + "/" + wikiAttachmentPath)},
		{"double leading slash", "attachment_path", inline("//" + wikiAttachmentPath)},
		{"missing slug", "attachment_path", inline(".files/report.data")},
		{"missing filename", "attachment_path", inline("team/.files/")},
		{"traversal segment", "attachment_path", inline("team/../.files/report.data")},
		{"query", "attachment_path", inline(wikiAttachmentPath + "?query=true")},
		{"both modes", "exactly one", both},
		{"no mode", "exactly one", map[string]any{"attachment_path": wikiAttachmentPath, "get_content": false}},
		{
			"inline extension",
			"attachment_path filename extension is not allowed for get_content; allowed extensions: data",
			inline("team/.files/report.txt"),
		},
		{"disk extension", "extension", save(filepath.Join(dir, "report.data"))},
		{"relative destination", "absolute", save("report.payload")},
		{"outside allowed directory", "allowed directories", save(filepath.Join(outside, "report.payload"))},
		{"existing destination", "already exists", save(existing)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertWikiAttachmentError(t, callWikiAttachment(t, session, tc.args), tc.message)
			assert.Zero(t, hits.Load())
		})
	}
	t.Cleanup(func() { assert.NoFileExists(t, path) })
}

// TestWikiAttachmentTransportErrorIsSafe ensures token refresh errors expose no credentials.
func TestWikiAttachmentTransportErrorIsSafe(t *testing.T) {
	t.Parallel()
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(fixture.Close)
	token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
	gomock.InOrder(
		token.EXPECT().Token(gomock.Any(), false).Return(wikiFixtureToken, nil),
		token.EXPECT().Token(gomock.Any(), true).Return("", errors.New("fixture-debug "+wikiFixtureToken)),
	)
	session := connectWikiAttachmentClient(t, fixture.URL, t.TempDir(), wikiAttachmentCap, token)
	result := callWikiAttachment(t, session, map[string]any{"attachment_path": wikiAttachmentPath, "get_content": true})
	assertWikiAttachmentError(t, result, "request failed")
}

// TestWikiAttachmentInterruptedBody checks that an interrupted HTTP 200 download never reports success.
func TestWikiAttachmentInterruptedBody(t *testing.T) {
	t.Parallel()
	for _, save := range []bool{false, true} {
		t.Run(fmt.Sprintf("save=%t", save), func(t *testing.T) {
			t.Parallel()
			fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertWikiDownloadRequest(t, r, wikiFixtureToken)
				w.Header().Set("Content-Length", "7")
				_, _ = io.WriteString(w, "part")
			}))
			t.Cleanup(fixture.Close)
			token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
			token.EXPECT().Token(gomock.Any(), false).Return(wikiFixtureToken, nil)
			dir := t.TempDir()
			session := connectWikiAttachmentClient(t, fixture.URL, dir, wikiAttachmentCap, token)
			args := map[string]any{"attachment_path": wikiAttachmentPath}
			if save {
				args["save_path"] = filepath.Join(dir, "result.payload")
			} else {
				args["get_content"] = true
			}
			assertWikiAttachmentError(t, callWikiAttachment(t, session, args), "complete Wiki attachment")
		})
	}
}

// TestWikiAttachmentSucceedsAfterRefresh checks delivery of the response authorized with the refreshed token.
func TestWikiAttachmentSucceedsAfterRefresh(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			assertWikiDownloadRequest(t, r, wikiFixtureToken)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "fixture-debug")
			return
		}
		assertWikiDownloadRequest(t, r, wikiRefreshedToken)
		_, _ = io.WriteString(w, "report")
	}))
	t.Cleanup(fixture.Close)
	token := apihelpers.NewMockITokenProvider(gomock.NewController(t))
	gomock.InOrder(
		token.EXPECT().Token(gomock.Any(), false).Return(wikiFixtureToken, nil),
		token.EXPECT().Token(gomock.Any(), true).Return(wikiRefreshedToken, nil),
	)
	session := connectWikiAttachmentClient(t, fixture.URL, t.TempDir(), wikiAttachmentCap, token)
	args := map[string]any{"attachment_path": wikiAttachmentPath, "get_content": true}
	result := callWikiAttachment(t, session, args)
	require.False(t, result.IsError, "%+v", result.Content)
	assert.Equal(t, map[string]any{"size": float64(6), "content": "report"}, result.StructuredContent)
	assert.EqualValues(t, 2, hits.Load())
}
