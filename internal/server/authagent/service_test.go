package authagent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// requestTimeout turns a request that can no longer complete into a prompt test failure.
const requestTimeout = 5 * time.Second

// newTestService builds a service whose sources fail the test on any call it does not expect.
func newTestService(t *testing.T) (service *Service, iam, oauth *MockITokenSource) {
	t.Helper()
	ctrl := gomock.NewController(t)
	iam, oauth = NewMockITokenSource(ctrl), NewMockITokenSource(ctrl)
	return New(iam, oauth), iam, oauth
}

// post sends one token request and returns the recorded response.
func post(ctx context.Context, service *Service, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	service.ServeHTTP(
		response,
		httptest.NewRequestWithContext(ctx, http.MethodPost, "/token", strings.NewReader(body)),
	)
	return response
}

// Protocol validation rejects malformed bodies, missing or unsupported token types, and missing
// profiles before invoking any source. A valid request reaches only the source of the requested
// token type and returns its token or its diagnostic: the other source is never substituted.
func TestTokenProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, body string
		status     int
		want       string
		tokenType  string
		profile    string
		failure    string
	}{
		{"iam success", `{"token_type":"iam","profile":"work"}`, 200, `"token":"iam-token"`, "iam", "work", ""},
		{"oauth success", `{"token_type":"oauth","profile":"work"}`, 200, `"token":"oauth-token"`, "oauth", "work", ""},
		{"iam profile", `{"token_type":"iam","profile":"other"}`, 200, `"token":"iam-token"`, "iam", "other", ""},
		{
			"oauth profile",
			`{"token_type":"oauth","profile":"other"}`,
			200,
			`"token":"oauth-token"`,
			"oauth",
			"other",
			"",
		},
		{
			"iam failure",
			`{"token_type":"iam","profile":"work"}`,
			502,
			`"message":"private-login-url"`,
			"iam",
			"work",
			"private-login-url",
		},
		{
			"oauth failure",
			`{"token_type":"oauth","profile":"absent"}`,
			502,
			`"message":"unknown OAuth profile"`,
			"oauth",
			"absent",
			"unknown OAuth profile",
		},
		{"missing token type", `{"profile":"work"}`, 400, `"message":"token_type is required"`, "", "", ""},
		{
			"empty token type",
			`{"token_type":"","profile":"work"}`,
			400,
			`"message":"token_type is required"`,
			"",
			"",
			"",
		},
		{"unsupported token type", `{"token_type":"jwt","profile":"work"}`, 400, `unsupported token_type`, "", "", ""},
		{"token type case", `{"token_type":"OAuth","profile":"work"}`, 400, `unsupported token_type`, "", "", ""},
		{"missing profile", `{"token_type":"oauth"}`, 400, `"message":"profile is required"`, "", "", ""},
		{"json", `{`, 400, `"message":`, "", "", ""},
		{
			"unknown member",
			`{"token_type":"iam","profile":"work","command":"anything"}`,
			400,
			`"message":`,
			"",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service, iam, oauth := newTestService(t)
			defer service.Close()
			requested := map[string]*MockITokenSource{"iam": iam, "oauth": oauth}[tt.tokenType]
			switch {
			case requested == nil:
			case tt.failure != "":
				requested.EXPECT().Acquire(gomock.Any(), tt.profile).Return("", errors.New(tt.failure))
			default:
				requested.EXPECT().Acquire(gomock.Any(), tt.profile).Return(tt.tokenType+"-token", nil)
			}
			// A source invoked against expectations stops its goroutine, so bound the request.
			ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
			defer cancel()
			response := post(ctx, service, tt.body)
			require.Equal(t, tt.status, response.Code)
			require.Contains(t, response.Body.String(), tt.want)
			if tt.status != http.StatusOK {
				require.NotContains(t, response.Body.String(), `"token"`)
			}
		})
	}
}

// IAM profile "work" and OAuth profile "work" identify different tokens. Concurrent requests
// for them run two acquisitions, one per source, and each caller receives its own token.
func TestTokenTypesDoNotShareAcquisitions(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		service, iam, oauth := newTestService(t)
		defer service.Close()
		release := make(chan struct{})
		var running atomic.Int32
		hold := func(token string) func(context.Context, string) (string, error) {
			return func(ctx context.Context, _ string) (string, error) {
				running.Add(1)
				select {
				case <-release:
					return token, nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
		}
		iam.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(hold("iam-token"))
		oauth.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(hold("oauth-token"))
		responses := make(chan [2]string, 2)
		for _, tokenType := range []string{"iam", "oauth"} {
			go func() {
				response := post(t.Context(), service, `{"token_type":"`+tokenType+`","profile":"work"}`)
				responses <- [2]string{tokenType, response.Body.String()}
			}()
		}
		synctest.Wait()
		assert.Equal(t, int32(2), running.Load())
		close(release)
		synctest.Wait()
		for range 2 {
			response := <-responses
			assert.Contains(t, response[1], `"token":"`+response[0]+`-token"`)
		}
	})
}

// Separate requests share one source acquisition. Canceling the first leaves
// the second running. Canceling the last stops the source. No network is used.
func TestRequestCancellation(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		service, source, _ := newTestService(t)
		defer service.Close()
		stopped := false
		source.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done()
			stopped = true
			return "", ctx.Err()
		})
		first, cancelFirst := context.WithCancel(t.Context())
		defer cancelFirst()
		second, cancelSecond := context.WithCancel(t.Context())
		defer cancelSecond()
		request := func(ctx context.Context) { post(ctx, service, `{"token_type":"iam","profile":"work"}`) }
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
