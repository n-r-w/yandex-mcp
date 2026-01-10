package apihelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// ErrorParseFunc is a function that parses an HTTP error into a domain error.
type ErrorParseFunc func(ctx context.Context, statusCode int, body []byte, operation string) error

// APIClient provides shared HTTP request methods for Yandex API adapters.
type APIClient struct {
	httpClient    *http.Client
	tokenProvider ITokenProvider
	orgID         string
	extraHeaders  map[string]string
	serviceName   string
	parseError    ErrorParseFunc
}

// APIClientConfig contains configuration for creating an APIClient.
type APIClientConfig struct {
	HTTPClient    *http.Client
	TokenProvider ITokenProvider
	OrgID         string
	ExtraHeaders  map[string]string
	ServiceName   string
	ParseError    ErrorParseFunc
	HTTPTimeout   time.Duration
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
	return &APIClient{
		httpClient:    httpClient,
		tokenProvider: cfg.TokenProvider,
		orgID:         cfg.OrgID,
		extraHeaders:  cfg.ExtraHeaders,
		serviceName:   cfg.ServiceName,
		parseError:    cfg.ParseError,
	}
}

// DoRequest performs an HTTP request and returns response headers.
func (c *APIClient) DoRequest(
	ctx context.Context,
	method, urlStr string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	resp, err := c.executeHTTPRequest(ctx, method, urlStr, body)
	if err != nil {
		return nil, c.ErrorLogWrapper(ctx, err)
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.Header, c.handleResponse(ctx, resp, result, operation)
}

// DoGET executes a GET request with token injection.
func (c *APIClient) DoGET(ctx context.Context, urlStr string, result any, operation string) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodGet, urlStr, nil, result, operation)
}

// DoPOST executes a POST request with token injection.
func (c *APIClient) DoPOST(
	ctx context.Context,
	urlStr string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodPost, urlStr, body, result, operation)
}

// DoPATCH executes a PATCH request with token injection.
func (c *APIClient) DoPATCH(
	ctx context.Context,
	urlStr string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodPatch, urlStr, body, result, operation)
}

// DoDELETE executes a DELETE request with token injection.
func (c *APIClient) DoDELETE(ctx context.Context, urlStr string, operation string) (http.Header, error) {
	return c.DoRequest(ctx, http.MethodDelete, urlStr, nil, nil, operation)
}

// handleResponse processes the HTTP response and decodes the result.
func (c *APIClient) handleResponse(ctx context.Context, resp *http.Response, result any, operation string) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.ErrorLogWrapper(ctx, fmt.Errorf("read response body: %w", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httpErr := &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       bodyBytes,
		}
		if c.parseError != nil {
			return c.parseError(ctx, httpErr.StatusCode, httpErr.Body, operation)
		}
		return c.ErrorLogWrapper(ctx, httpErr)
	}

	if result != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return c.ErrorLogWrapper(ctx, fmt.Errorf("decode response: %w", err))
		}
	}

	return nil
}

// executeHTTPRequest performs a single HTTP request with token injection and optional body encoding.
func (c *APIClient) executeHTTPRequest(
	ctx context.Context,
	method, urlStr string,
	body any,
) (*http.Response, error) {
	token, err := c.tokenProvider.Token(ctx)
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

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set(HeaderAuthorization, "Bearer "+token)
	req.Header.Set(HeaderCloudOrgID, c.orgID)
	req.Header.Set(HeaderContentType, ContentTypeJSON)

	for key, value := range c.extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
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
