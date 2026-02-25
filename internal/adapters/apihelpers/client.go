package apihelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// ErrorParseFunc is a function that parses an HTTP error into a domain error.
type ErrorParseFunc func(ctx context.Context, statusCode int, body []byte, operation string) error

// APIClient provides shared HTTP request methods for Yandex API adapters.
type APIClient struct {
	httpDoer            IHTTPDoer
	tokenProvider       ITokenProvider
	baseURL             *url.URL
	baseURLParseErr     error
	orgID               string
	extraHeaders        map[string]string
	serviceName         string
	parseError          ErrorParseFunc
	rawResponseMaxBytes int64
}

// APIClientConfig contains configuration for creating an APIClient.
type APIClientConfig struct {
	HTTPClient          *http.Client
	TokenProvider       ITokenProvider
	BaseURL             string
	OrgID               string
	ExtraHeaders        map[string]string
	ServiceName         string
	ParseError          ErrorParseFunc
	HTTPTimeout         time.Duration
	RawResponseMaxBytes int64
}

// NewAPIClient creates a new APIClient with the given configuration.
func NewAPIClient(cfg APIClientConfig) *APIClient {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.HTTPTimeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		httpClient = &http.Client{ //nolint:exhaustruct // optional fields use defaults
			Timeout: timeout,
		}
	}

	parsedBaseURL, baseURLParseErr := parseBaseURL(cfg.BaseURL)

	return &APIClient{
		httpDoer:            httpClient,
		tokenProvider:       cfg.TokenProvider,
		baseURL:             parsedBaseURL,
		baseURLParseErr:     baseURLParseErr,
		orgID:               cfg.OrgID,
		extraHeaders:        cfg.ExtraHeaders,
		serviceName:         cfg.ServiceName,
		parseError:          cfg.ParseError,
		rawResponseMaxBytes: cfg.RawResponseMaxBytes,
	}
}

// DoRequest performs an HTTP request and returns response headers.
func (c *APIClient) DoRequest(
	ctx context.Context,
	method, endpointPath string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	resp, err := c.executeRequestWithAuthRetry(ctx, method, endpointPath, body)
	if err != nil {
		return nil, c.ErrorLogWrapper(ctx, err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close response body", "error", err)
		}
	}()

	return resp.Header, c.handleResponse(ctx, resp, result, operation)
}

// DoRequestRaw performs an HTTP request and returns response headers and body bytes.
func (c *APIClient) DoRequestRaw(
	ctx context.Context,
	method, endpointPath string,
	body any,
	operation string,
) (http.Header, []byte, error) {
	resp, err := c.executeRequestWithAuthRetry(ctx, method, endpointPath, body)
	if err != nil {
		return nil, nil, c.ErrorLogWrapper(ctx, err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close response body", "error", err)
		}
	}()

	bodyBytes, err := readResponseBody(resp.Body, c.rawResponseMaxBytes)
	if err != nil {
		return nil, nil, c.ErrorLogWrapper(ctx, fmt.Errorf("read response body: %w", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, c.parseHTTPError(ctx, resp.StatusCode, bodyBytes, operation)
	}

	return resp.Header, bodyBytes, nil
}

// DoRequestStream performs an HTTP request and returns response headers and body stream.
// Caller is responsible for closing the returned body.
func (c *APIClient) DoRequestStream(
	ctx context.Context,
	method, endpointPath string,
	body any,
	operation string,
) (http.Header, io.ReadCloser, error) {
	resp, err := c.executeRequestWithAuthRetry(ctx, method, endpointPath, body)
	if err != nil {
		return nil, nil, c.ErrorLogWrapper(ctx, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, readErr := readResponseBody(resp.Body, c.rawResponseMaxBytes)
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
		if readErr != nil {
			return nil, nil, c.ErrorLogWrapper(ctx, fmt.Errorf("read response body: %w", readErr))
		}
		return nil, nil, c.parseHTTPError(ctx, resp.StatusCode, bodyBytes, operation)
	}

	return resp.Header, resp.Body, nil
}

// readResponseBody reads response body with an optional size limit.
func readResponseBody(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(reader)
	}
	limited := &io.LimitedReader{R: reader, N: maxBytes + 1}
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds max size of %d bytes", maxBytes)
	}
	return body, nil
}

// DoGET executes a GET request with token injection.
func (c *APIClient) DoGET(ctx context.Context, endpointPath string, result any, operation string) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodGet, endpointPath, nil, result, operation)
}

// DoGETRaw executes a GET request and returns response headers and body bytes.
func (c *APIClient) DoGETRaw(ctx context.Context, endpointPath string, operation string) (http.Header, []byte, error) {
	return c.DoRequestRaw(ctx, http.MethodGet, endpointPath, nil, operation)
}

// DoGETStream executes a GET request and returns response headers and body stream.
// Caller is responsible for closing the returned body.
func (c *APIClient) DoGETStream(
	ctx context.Context,
	endpointPath string,
	operation string,
) (http.Header, io.ReadCloser, error) {
	return c.DoRequestStream(ctx, http.MethodGet, endpointPath, nil, operation)
}

// DoPOST executes a POST request with token injection.
func (c *APIClient) DoPOST(
	ctx context.Context,
	endpointPath string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodPost, endpointPath, body, result, operation)
}

// DoPATCH executes a PATCH request with token injection.
func (c *APIClient) DoPATCH(
	ctx context.Context,
	endpointPath string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodPatch, endpointPath, body, result, operation)
}

// DoDELETE executes a DELETE request with token injection.
func (c *APIClient) DoDELETE(ctx context.Context, endpointPath string, operation string) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodDelete, endpointPath, nil, nil, operation)
}

// handleResponse processes the HTTP response and decodes the result.
func (c *APIClient) handleResponse(ctx context.Context, resp *http.Response, result any, operation string) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.ErrorLogWrapper(ctx, fmt.Errorf("read response body: %w", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseHTTPError(ctx, resp.StatusCode, bodyBytes, operation)
	}

	if result != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return c.ErrorLogWrapper(ctx, fmt.Errorf("decode response: %w", err))
		}
	}

	return nil
}

// executeRequestWithAuthRetry performs a request and retries once after forced token refresh
// when the upstream responds with authentication/authorization errors.
func (c *APIClient) executeRequestWithAuthRetry(
	ctx context.Context,
	method, endpointPath string,
	body any,
) (*http.Response, error) {
	resp, err := c.executeHTTPRequest(ctx, method, endpointPath, body, false)
	if err != nil {
		return nil, err
	}

	if !isAuthRetryStatus(resp.StatusCode) {
		return resp, nil
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		slog.WarnContext(ctx, "failed to close response body before retry", "error", closeErr)
	}

	resp, err = c.executeHTTPRequest(ctx, method, endpointPath, body, true)
	if err != nil {
		return nil, fmt.Errorf("failed retry after token refresh: %w", err)
	}

	return resp, nil
}

// parseHTTPError converts raw HTTP status/body to either custom parsed error or generic HTTPError.
func (c *APIClient) parseHTTPError(ctx context.Context, statusCode int, body []byte, operation string) error {
	httpErr := &HTTPError{
		StatusCode: statusCode,
		Body:       body,
	}

	if c.parseError != nil {
		return c.parseError(ctx, httpErr.StatusCode, httpErr.Body, operation)
	}

	return c.ErrorLogWrapper(ctx, httpErr)
}

// isAuthRetryStatus reports whether status code should trigger a single token-refresh retry.
func isAuthRetryStatus(statusCode int) bool {
	return statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden
}

// parseBaseURL ensures API base URL is absolute and uses an allowed transport scheme.
func parseBaseURL(baseURL string) (*url.URL, error) {
	if baseURL == "" {
		return nil, errors.New("base URL must not be empty")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	if !parsedURL.IsAbs() {
		return nil, errors.New("base URL must be absolute")
	}

	if parsedURL.Hostname() == "" {
		return nil, errors.New("base URL must include host")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, errors.New("base URL scheme must be http or https")
	}

	path := strings.TrimSpace(parsedURL.Path)
	if path != "" && strings.Trim(path, "/") != "" {
		return nil, errors.New("base URL must not include path")
	}

	parsedURL.Path = ""

	return parsedURL, nil
}

// resolveRequestURL builds an absolute request URL from a trusted base URL and a relative endpoint path.
func (c *APIClient) resolveRequestURL(endpointPath string) (string, error) {
	if c.baseURLParseErr != nil {
		return "", fmt.Errorf("invalid API base URL: %w", c.baseURLParseErr)
	}

	if c.baseURL == nil {
		return "", errors.New("API base URL is not configured")
	}

	if endpointPath == "" {
		return "", errors.New("endpoint path must not be empty")
	}

	relativeURL, err := url.Parse(endpointPath)
	if err != nil {
		return "", fmt.Errorf("parse endpoint path: %w", err)
	}

	if relativeURL.IsAbs() || relativeURL.Host != "" {
		return "", errors.New("endpoint path must be relative")
	}

	if !strings.HasPrefix(relativeURL.Path, "/") {
		return "", errors.New("endpoint path must start with '/'")
	}

	resolvedURL := c.baseURL.ResolveReference(relativeURL)
	if resolvedURL.Scheme != c.baseURL.Scheme || resolvedURL.Host != c.baseURL.Host {
		return "", errors.New("resolved endpoint escaped configured base URL")
	}

	return resolvedURL.String(), nil
}

// executeHTTPRequest performs a single HTTP request with token injection and optional body encoding.
func (c *APIClient) executeHTTPRequest(
	ctx context.Context,
	method, endpointPath string,
	body any,
	tokenForceRefresh bool,
) (*http.Response, error) {
	requestURL, err := c.resolveRequestURL(endpointPath)
	if err != nil {
		return nil, fmt.Errorf("resolve request URL: %w", err)
	}

	token, err := c.tokenProvider.Token(ctx, tokenForceRefresh)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal request body: %w", marshalErr)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set(HeaderAuthorization, "Bearer "+token)
	req.Header.Set(HeaderCloudOrgID, c.orgID)
	req.Header.Set(HeaderContentType, ContentTypeJSON)

	for key, value := range c.extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	return resp, nil
}

// ErrorLogWrapper logs an error with the service name prefix and returns it.
// Useful for errors that occur outside the APIClient methods.
func (c *APIClient) ErrorLogWrapper(ctx context.Context, err error) error {
	return domain.LogError(ctx, c.serviceName+" adapter error", err)
}
