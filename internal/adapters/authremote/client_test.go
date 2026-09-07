package authremote

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Protocol contract with a real loopback server. Inputs cover success, all safe
// error codes, incompatible/malformed replies, and redirects. Replies must not leak.
func TestAcquire(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		status     int
		body, want string
	}{
		{"success", 200, `{"version":1,"token":"token"}`, ""},
		{
			"forbidden",
			403,
			`{"version":1,"error":"forbidden_profile","message":"profile access denied"}`,
			"profile access denied",
		},
		{
			"authentication",
			502,
			`{"version":1,"error":"authentication_failed","message":"exit status 1: diagnostic text"}`,
			"exit status 1: diagnostic text",
		},
		{
			"version",
			400,
			`{"version":1,"error":"incompatible_version","message":"version not supported"}`,
			"incompatible auth-agent protocol",
		},
		{
			"invalid",
			400,
			`{"version":1,"error":"invalid_request","message":"profile missing"}`,
			"invalid token request",
		},
		{
			"internal",
			500,
			`{"version":1,"error":"internal_error","message":"internal failure"}`,
			"auth-agent internal error",
		},
		{"malformed", 200, `private-login-url`, "invalid character"},
		{"unknown", 200, `{"version":1,"token":"token","other":"private-login-url"}`, "invalid auth-agent response"},
		{"empty", 200, `{"version":1}`, "invalid auth-agent response"},
		{"contradictory", 200, `{"version":1,"token":"token","error":"internal_error"}`, "invalid auth-agent response"},
		{"wrongVersion", 200, `{"version":2,"token":"token"}`, "incompatible auth-agent protocol"},
		{"wrongStatus", 200, `{"version":1,"error":"authentication_failed"}`, "invalid auth-agent response"},
		{
			"longError",
			502,
			`{"version":1,"error":"authentication_failed","message":"` + strings.Repeat("x", 5000) + `"}`,
			strings.Repeat("x", 5000),
		},
		{"redirect", 302, `private-login-url`, "invalid auth-agent response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/token", r.URL.Path)
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, `{"version":1,"profile":"work"}`, string(body))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Location", "http://127.0.0.1:1/private-login-url")
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			source, err := New(srv.URL)
			require.NoError(t, err)
			defer source.Close()
			token, err := source.Acquire(t.Context(), "work")
			if tt.want == "" {
				require.NoError(t, err)
				require.Equal(t, "token", token)
			} else {
				require.ErrorContains(t, err, tt.want)
				require.Empty(t, token)
				require.NotContains(t, err.Error(), "private-login-url")
			}
		})
	}
}

// Only explicit loopback HTTP endpoints are valid, with no credentials or query.
func TestEndpointValidation(t *testing.T) {
	t.Parallel()
	for _, endpoint := range []string{
		"https://127.0.0.1:18765", "http://localhost:18765", "http://127.0.0.1",
		"http://127.0.0.1:0", "http://127.0.0.1:99999", "http://user@127.0.0.1:1",
		"http://127.0.0.1:1/path", "http://127.0.0.1:1?query", "http://127.0.0.1:1#fragment",
	} {
		_, err := New(endpoint)
		require.Error(t, err, endpoint)
	}
}

// Closing the caller's wait cancels the HTTP request on the workstation.
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
	source, err := New(srv.URL)
	require.NoError(t, err)
	defer source.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, requestErr := source.Acquire(ctx, "work"); result <- requestErr }()
	<-started
	cancel()
	require.ErrorIs(t, <-result, context.Canceled)
	<-stopped
}
