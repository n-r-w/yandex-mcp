package authagent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Protocol validation rejects invalid versions, bodies, and forbidden profiles
// before invoking the source mock. Valid requests receive a token or safe error.
func TestTokenProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, body string
		status     int
		code       string
		invoke     bool
		failure    bool
	}{
		{"success", `{"version":1,"profile":"work"}`, 200, `"token":"token"`, true, false},
		{"failure", `{"version":1,"profile":"work"}`, 502, `"error":"authentication_failed"`, true, true},
		{"version", `{"version":2,"profile":"work"}`, 400, `"error":"incompatible_version"`, false, false},
		{"profile", `{"version":1,"profile":"other"}`, 403, `"error":"forbidden_profile"`, false, false},
		{"missing", `{"version":1}`, 400, `"error":"invalid_request"`, false, false},
		{"json", `{`, 400, `"error":"invalid_request"`, false, false},
		{
			"unknown",
			`{"version":1,"profile":"work","command":"anything"}`,
			400,
			`"error":"invalid_request"`,
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewMockITokenSource(gomock.NewController(t))
			if tt.invoke {
				if tt.failure {
					source.EXPECT().Acquire(gomock.Any(), "work").Return("", errors.New("private-login-url"))
				} else {
					source.EXPECT().Acquire(gomock.Any(), "work").Return("token", nil)
				}
			}
			service := New(source, []string{"work"})
			defer service.Close()
			response := httptest.NewRecorder()
			service.ServeHTTP(
				response,
				httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/token", strings.NewReader(tt.body)),
			)
			require.Equal(t, tt.status, response.Code)
			require.Contains(t, response.Body.String(), tt.code)
			if tt.failure {
				require.Contains(t, response.Body.String(), `"message":"private-login-url"`)
			}
		})
	}
}

// Separate requests share one source acquisition. Canceling the first leaves
// the second running. Canceling the last stops the source. No network is used.
func TestRequestCancellation(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		source := NewMockITokenSource(gomock.NewController(t))
		stopped := false
		source.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done()
			stopped = true
			return "", ctx.Err()
		})
		service := New(source, []string{"work"})
		defer service.Close()
		first, cancelFirst := context.WithCancel(t.Context())
		defer cancelFirst()
		second, cancelSecond := context.WithCancel(t.Context())
		defer cancelSecond()
		request := func(ctx context.Context) {
			service.ServeHTTP(
				httptest.NewRecorder(),
				httptest.NewRequestWithContext(
					ctx,
					http.MethodPost,
					"/token",
					strings.NewReader(`{"version":1,"profile":"work"}`),
				),
			)
		}
		go request(first)
		synctest.Wait()
		go request(second)
		synctest.Wait()
		cancelFirst()
		synctest.Wait()
		assert.False(t, stopped)
		cancelSecond()
		synctest.Wait()
		assert.True(t, stopped)
	})
}
