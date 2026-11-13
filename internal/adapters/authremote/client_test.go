package authremote

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcquire checks token exchange, original diagnostics, and malformed responses.
func TestAcquire(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		status     int
		body, want string
	}{
		{"success", 200, `{"token":"token"}`, ""},
		{
			"authentication",
			502,
			`{"message":"exit status 1: original command diagnostic"}`,
			"exit status 1: original command diagnostic",
		},
		{"other HTTP error", 503, `{"message":"source unavailable"}`, "source unavailable"},
		{"long diagnostic", 502, `{"message":"` + strings.Repeat("x", 5000) + `"}`, strings.Repeat("x", 5000)},
		{"malformed", 200, `not JSON`, "invalid character"},
		{"empty token", 200, `{}`, "invalid auth-agent response"},
		{"conflicting fields", 200, `{"token":"token","message":"failed"}`, "invalid auth-agent response"},
		{"redirect", 302, `{}`, "invalid auth-agent response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/token", r.URL.Path)
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, `{"token_type":"iam","profile":"work"}`, string(body))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Location", "http://127.0.0.1:1/redirect-target")
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			source, err := New(srv.Listener.Addr().(*net.TCPAddr).Port, "iam")
			require.NoError(t, err)
			defer source.Close()
			token, err := source.Acquire(t.Context(), "work")
			if tt.want == "" {
				require.NoError(t, err)
				require.Equal(t, "token", token)
			} else {
				require.ErrorContains(t, err, tt.want)
				require.Empty(t, token)
				if tt.status >= 400 {
					require.ErrorContains(t, err, fmt.Sprintf("HTTP %d", tt.status))
				}
			}
		})
	}
}

// TestTokenTypeInRequest sends the configured token type with the profile and refuses
// to send a request that names no token type.
func TestTokenTypeInRequest(t *testing.T) {
	t.Parallel()
	requests := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		requests <- string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"token":"oauth-token"}`)
	}))
	defer srv.Close()
	port := srv.Listener.Addr().(*net.TCPAddr).Port

	source, err := New(port, "oauth")
	require.NoError(t, err)
	defer source.Close()
	token, err := source.Acquire(t.Context(), "work")
	require.NoError(t, err)
	require.Equal(t, "oauth-token", token)
	require.JSONEq(t, `{"token_type":"oauth","profile":"work"}`, <-requests)

	untyped, err := New(port, "")
	require.NoError(t, err)
	defer untyped.Close()
	token, err = untyped.Acquire(t.Context(), "work")
	require.ErrorContains(t, err, "invalid token request")
	require.Empty(t, token)
	require.Empty(t, requests)
}

// TestPortValidation rejects ports outside the TCP range.
func TestPortValidation(t *testing.T) {
	t.Parallel()
	for _, port := range []int{0, -1, 65536} {
		_, err := New(port, "iam")
		require.ErrorContains(t, err, "YANDEX_MCP_AUTH_AGENT_PORT")
	}
}

// TestCancellation verifies that leaving a request cancels the workstation wait.
func TestCancellation(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	stopped := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	defer srv.Close()
	source, err := New(srv.Listener.Addr().(*net.TCPAddr).Port, "iam")
	require.NoError(t, err)
	defer source.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := source.Acquire(ctx, "work"); result <- err }()
	<-started
	cancel()
	require.ErrorIs(t, <-result, context.Canceled)
	<-stopped
}
