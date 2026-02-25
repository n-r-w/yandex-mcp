package apihelpers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newTestAPIClient creates an API client with injected fake dependencies.
func newTestAPIClient(doer IHTTPDoer, provider ITokenProvider) *APIClient {
	parsedBaseURL, err := url.Parse("https://api.example.test")
	if err != nil {
		panic(err)
	}

	return &APIClient{
		httpDoer:            doer,
		tokenProvider:       provider,
		baseURL:             parsedBaseURL,
		baseURLParseErr:     nil,
		orgID:               "org-id",
		extraHeaders:        nil,
		serviceName:         "test-service",
		parseError:          nil,
		rawResponseMaxBytes: 0,
	}
}

// TestParseBaseURL_Validation verifies base URL validation behavior.
func TestParseBaseURL_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{name: "valid https", baseURL: "https://api.example.test", wantErr: false},
		{name: "valid http", baseURL: "http://api.example.test", wantErr: false},
		{name: "empty", baseURL: "", wantErr: true},
		{name: "relative", baseURL: "/v1", wantErr: true},
		{name: "missing host", baseURL: "https:///v1", wantErr: true},
		{name: "unsupported scheme", baseURL: "ftp://api.example.test", wantErr: true},
		{name: "base URL with path is rejected", baseURL: "https://api.example.test/v1", wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parsedURL, err := parseBaseURL(testCase.baseURL)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, parsedURL)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, parsedURL)
		})
	}

	t.Run("base URL parse error is returned", func(t *testing.T) {
		t.Parallel()

		_, err := parseBaseURL("://bad")
		require.Error(t, err)
	})
}

// TestResolveRequestURL_Validation verifies endpoint path resolution and host-lock checks.
func TestResolveRequestURL_Validation(t *testing.T) {
	t.Parallel()

	parsedBaseURL, err := url.Parse("https://api.example.test")
	require.NoError(t, err)

	client := &APIClient{
		httpDoer:            nil,
		tokenProvider:       nil,
		baseURL:             parsedBaseURL,
		baseURLParseErr:     nil,
		orgID:               "",
		extraHeaders:        nil,
		serviceName:         "",
		parseError:          nil,
		rawResponseMaxBytes: 0,
	}

	testCases := []struct {
		name         string
		endpointPath string
		wantErr      bool
		wantURL      string
	}{
		{name: "valid relative endpoint", endpointPath: "/v1/pages?x=1", wantErr: false, wantURL: "https://api.example.test/v1/pages?x=1"},
		{name: "empty endpoint", endpointPath: "", wantErr: true, wantURL: ""},
		{name: "absolute endpoint", endpointPath: "https://evil.test/p", wantErr: true, wantURL: ""},
		{name: "host escape endpoint", endpointPath: "//evil.test/p", wantErr: true, wantURL: ""},
		{name: "missing leading slash", endpointPath: "v1/pages", wantErr: true, wantURL: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resolvedURL, resolveErr := client.resolveRequestURL(testCase.endpointPath)
			if testCase.wantErr {
				require.Error(t, resolveErr)
				return
			}

			require.NoError(t, resolveErr)
			assert.Equal(t, testCase.wantURL, resolvedURL)
		})
	}

	t.Run("base URL parse error in client state", func(t *testing.T) {
		t.Parallel()

		clientWithParseError := &APIClient{
			httpDoer:            nil,
			tokenProvider:       nil,
			baseURL:             parsedBaseURL,
			baseURLParseErr:     errors.New("invalid base URL"),
			orgID:               "",
			extraHeaders:        nil,
			serviceName:         "",
			parseError:          nil,
			rawResponseMaxBytes: 0,
		}

		_, err := clientWithParseError.resolveRequestURL("/v1/pages")
		require.Error(t, err)
	})

	t.Run("nil base URL in client state", func(t *testing.T) {
		t.Parallel()

		clientWithNilBase := &APIClient{
			httpDoer:            nil,
			tokenProvider:       nil,
			baseURL:             nil,
			baseURLParseErr:     nil,
			orgID:               "",
			extraHeaders:        nil,
			serviceName:         "",
			parseError:          nil,
			rawResponseMaxBytes: 0,
		}

		_, err := clientWithNilBase.resolveRequestURL("/v1/pages")
		require.Error(t, err)
	})
}

// TestDoGETRaw_RetriesAfterUnauthorized verifies auth retry for raw responses.
func TestDoGETRaw_RetriesAfterUnauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	doer := NewMockIHTTPDoer(ctrl)
	provider := NewMockITokenProvider(ctrl)

	firstResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
		Header:     make(http.Header),
	}
	secondResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
		Header:     make(http.Header),
	}

	gomock.InOrder(
		provider.EXPECT().Token(gomock.Any(), false).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(firstResponse, nil),
		provider.EXPECT().Token(gomock.Any(), true).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(secondResponse, nil),
	)

	client := newTestAPIClient(doer, provider)

	headers, body, err := client.DoGETRaw(t.Context(), "/v1/resource", "operation")
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, []byte("ok"), body)
}

// TestDoGETRaw_DoesNotRetryOnNonAuthStatus verifies non-auth statuses do not trigger token-refresh retry.
func TestDoGETRaw_DoesNotRetryOnNonAuthStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	doer := NewMockIHTTPDoer(ctrl)
	provider := NewMockITokenProvider(ctrl)

	errorResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("upstream-error")),
		Header:     make(http.Header),
	}

	provider.EXPECT().Token(gomock.Any(), false).Return("token", nil)
	doer.EXPECT().Do(gomock.Any()).Return(errorResponse, nil)

	client := newTestAPIClient(doer, provider)

	_, _, err := client.DoGETRaw(t.Context(), "/v1/resource", "operation")
	require.Error(t, err)
}

// TestDoGETStream_RetriesAfterForbidden verifies auth retry for stream responses.
func TestDoGETStream_RetriesAfterForbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	doer := NewMockIHTTPDoer(ctrl)
	provider := NewMockITokenProvider(ctrl)

	firstResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(bytes.NewBufferString("forbidden")),
		Header:     make(http.Header),
	}
	secondResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("stream-body")),
		Header:     make(http.Header),
	}

	gomock.InOrder(
		provider.EXPECT().Token(gomock.Any(), false).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(firstResponse, nil),
		provider.EXPECT().Token(gomock.Any(), true).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(secondResponse, nil),
	)

	client := newTestAPIClient(doer, provider)

	headers, stream, err := client.DoGETStream(t.Context(), "/v1/stream", "operation")
	require.NoError(t, err)
	require.NotNil(t, headers)
	require.NotNil(t, stream)
	t.Cleanup(func() {
		require.NoError(t, stream.Close())
	})

	streamBody, readErr := io.ReadAll(stream)
	require.NoError(t, readErr)
	assert.Equal(t, []byte("stream-body"), streamBody)
}

// TestDoGETStream_RetryFailureIsWrapped verifies second-attempt request errors are wrapped with retry context.
func TestDoGETStream_RetryFailureIsWrapped(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	doer := NewMockIHTTPDoer(ctrl)
	provider := NewMockITokenProvider(ctrl)

	firstResponse := &http.Response{ //nolint:exhaustruct // optional http.Response fields are irrelevant for this test case
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
		Header:     make(http.Header),
	}

	retryErr := errors.New("network down")

	gomock.InOrder(
		provider.EXPECT().Token(gomock.Any(), false).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(firstResponse, nil),
		provider.EXPECT().Token(gomock.Any(), true).Return("token", nil),
		doer.EXPECT().Do(gomock.Any()).Return(nil, retryErr),
	)

	client := newTestAPIClient(doer, provider)

	_, _, err := client.DoGETStream(t.Context(), "/v1/stream", "operation")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed retry after token refresh")
}
