// Package tracker provides HTTP client for Yandex Tracker API.
package tracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	trackertools "github.com/n-r-w/yandex-mcp/internal/tools/tracker"
)

// Client implements HTTP client for Yandex Tracker API.
type Client struct {
	httpClient    *http.Client
	tokenProvider ITokenProvider
	baseURL       string
	orgID         string
}

// Compile-time check that Client implements the tools interface.
var _ trackertools.ITrackerAdapter = (*Client)(nil)

// NewClient creates a new Tracker API client.
func NewClient(cfg *config.Config, tokenProvider ITokenProvider) *Client {
	return &Client{
		httpClient: &http.Client{ //nolint:exhaustruct // optional fields use defaults
			Timeout: defaultTimeout,
		},
		tokenProvider: tokenProvider,
		baseURL:       strings.TrimSuffix(cfg.TrackerBaseURL, "/"),
		orgID:         cfg.CloudOrgID,
	}
}

// GetIssue retrieves an issue by its ID or key.
func (c *Client) GetIssue(
	ctx context.Context,
	issueID string,
	opts domain.TrackerGetIssueOpts,
) (*domain.TrackerIssue, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	if opts.Expand != "" {
		q := u.Query()
		q.Set("expand", opts.Expand)
		u.RawQuery = q.Encode()
	}

	var issue Issue
	if _, err := c.doGET(ctx, u.String(), &issue, "GetIssue"); err != nil {
		return nil, err
	}
	result := issueToTrackerIssue(issue)
	return &result, nil
}

// SearchIssues searches for issues using filter or query.
func (c *Client) SearchIssues(
	ctx context.Context,
	opts domain.TrackerSearchIssuesOpts,
) (*domain.TrackerIssuesPage, error) {
	u, err := url.Parse(c.baseURL + "/v3/issues/_search")
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if opts.Expand != "" {
		q.Set("expand", opts.Expand)
	}

	// Standard pagination parameters
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}

	// Scroll pagination parameters
	if opts.ScrollType != "" {
		q.Set("scrollType", opts.ScrollType)
	}
	if opts.PerScroll > 0 {
		q.Set("perScroll", strconv.Itoa(opts.PerScroll))
	}
	if opts.ScrollTTLMillis > 0 {
		q.Set("scrollTTLMillis", strconv.Itoa(opts.ScrollTTLMillis))
	}
	if opts.ScrollID != "" {
		q.Set("scrollId", opts.ScrollID)
	}
	u.RawQuery = q.Encode()

	reqBody := searchRequest{
		Filter: stringMapToAnyMap(opts.Filter),
		Query:  opts.Query,
		Order:  opts.Order,
	}

	var issues []Issue
	headers, err := c.doPOST(ctx, u.String(), reqBody, &issues, "SearchIssues")
	if err != nil {
		return nil, err
	}

	dtoResult := SearchIssuesResult{
		Issues:      issues,
		TotalCount:  parseIntHeaderValue(headers, headerXTotalCount),
		TotalPages:  parseIntHeaderValue(headers, headerXTotalPages),
		ScrollID:    headers.Get(headerXScrollID),
		ScrollToken: headers.Get(headerXScrollToken),
		NextLink:    headers.Get(headerLink),
	}
	result := searchIssuesResultToTrackerIssuesPage(dtoResult)
	return &result, nil
}

// CountIssues counts issues matching the filter or query.
func (c *Client) CountIssues(ctx context.Context, opts domain.TrackerCountIssuesOpts) (int, error) {
	u, err := url.Parse(c.baseURL + "/v3/issues/_count")
	if err != nil {
		return 0, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	reqBody := countRequest{
		Filter: stringMapToAnyMap(opts.Filter),
		Query:  opts.Query,
	}

	var count int
	if _, err := c.doPOST(ctx, u.String(), reqBody, &count, "CountIssues"); err != nil {
		return 0, err
	}

	return count, nil
}

// ListIssueTransitions lists available transitions for an issue.
func (c *Client) ListIssueTransitions(
	ctx context.Context,
	issueID string,
) ([]domain.TrackerTransition, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s/transitions", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	var transitions []Transition
	if _, err := c.doGET(ctx, u.String(), &transitions, "ListIssueTransitions"); err != nil {
		return nil, err
	}
	result := make([]domain.TrackerTransition, len(transitions))
	for i, t := range transitions {
		result[i] = transitionToTrackerTransition(t)
	}
	return result, nil
}

// ListQueues lists all queues.
func (c *Client) ListQueues(
	ctx context.Context,
	opts domain.TrackerListQueuesOpts,
) (*domain.TrackerQueuesPage, error) {
	u, err := url.Parse(c.baseURL + "/v3/queues/")
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if opts.Expand != "" {
		q.Set("expand", opts.Expand)
	}
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	u.RawQuery = q.Encode()

	var queues []Queue
	headers, err := c.doGET(ctx, u.String(), &queues, "ListQueues")
	if err != nil {
		return nil, err
	}

	dtoResult := ListQueuesResult{
		Queues:     queues,
		TotalCount: parseIntHeaderValue(headers, headerXTotalCount),
		TotalPages: parseIntHeaderValue(headers, headerXTotalPages),
	}
	result := listQueuesResultToTrackerQueuesPage(dtoResult)
	return &result, nil
}

// ListIssueComments lists comments for an issue.
func (c *Client) ListIssueComments(
	ctx context.Context,
	issueID string,
	opts domain.TrackerListCommentsOpts,
) (*domain.TrackerCommentsPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s/comments", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if opts.Expand != "" {
		q.Set("expand", opts.Expand)
	}
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.ID > 0 {
		q.Set("id", strconv.FormatInt(opts.ID, 10))
	}
	u.RawQuery = q.Encode()

	var comments []Comment
	headers, err := c.doGET(ctx, u.String(), &comments, "ListIssueComments")
	if err != nil {
		return nil, err
	}

	dtoResult := ListCommentsResult{
		Comments: comments,
		NextLink: headers.Get(headerLink),
	}
	result := listCommentsResultToTrackerCommentsPage(dtoResult)
	return &result, nil
}

// CreateIssue creates a new Tracker issue.
func (c *Client) CreateIssue(
	ctx context.Context,
	req *domain.TrackerIssueCreateRequest,
) (*domain.TrackerIssueCreateResponse, error) {
	u := c.baseURL + "/v3/issues"

	body := CreateIssueRequest{
		Queue:         req.Queue,
		Summary:       req.Summary,
		Description:   req.Description,
		Type:          req.Type,
		Priority:      req.Priority,
		Assignee:      req.Assignee,
		Tags:          req.Tags,
		Parent:        req.Parent,
		AttachmentIDs: req.AttachmentIDs,
		Sprint:        req.Sprint,
	}

	var issue Issue
	if _, err := c.doPOST(ctx, u, body, &issue, "CreateIssue"); err != nil {
		return nil, err
	}

	result := issueToTrackerIssue(issue)
	return &domain.TrackerIssueCreateResponse{Issue: result}, nil
}

// UpdateIssue updates an existing Tracker issue.
func (c *Client) UpdateIssue(
	ctx context.Context,
	req *domain.TrackerIssueUpdateRequest,
) (*domain.TrackerIssueUpdateResponse, error) {
	u := fmt.Sprintf("%s/v3/issues/%s", c.baseURL, url.PathEscape(req.IssueID))

	body := UpdateIssueRequest{
		Summary:     req.Summary,
		Description: req.Description,
		Type:        req.Type,
		Priority:    req.Priority,
		Assignee:    req.Assignee,
		Version:     req.Version,
	}

	var issue Issue
	if _, err := c.doPATCH(ctx, u, body, &issue, "UpdateIssue"); err != nil {
		return nil, err
	}

	result := issueToTrackerIssue(issue)
	return &domain.TrackerIssueUpdateResponse{Issue: result}, nil
}

// ExecuteTransition executes a status transition on an issue.
func (c *Client) ExecuteTransition(
	ctx context.Context,
	req *domain.TrackerTransitionExecuteRequest,
) (*domain.TrackerTransitionExecuteResponse, error) {
	u := fmt.Sprintf("%s/v3/issues/%s/transitions/%s/_execute",
		c.baseURL,
		url.PathEscape(req.IssueID),
		url.PathEscape(req.TransitionID),
	)

	var body any
	if req.Comment != "" || len(req.Fields) > 0 {
		body = ExecuteTransitionRequest{
			Comment: req.Comment,
			Fields:  req.Fields,
		}
	}

	var transitions []Transition
	if _, err := c.doPOST(ctx, u, body, &transitions, "ExecuteTransition"); err != nil {
		return nil, err
	}

	result := make([]domain.TrackerTransition, len(transitions))
	for i, t := range transitions {
		result[i] = transitionToTrackerTransition(t)
	}

	return &domain.TrackerTransitionExecuteResponse{Transitions: result}, nil
}

// AddComment adds a comment to an issue.
func (c *Client) AddComment(
	ctx context.Context,
	req *domain.TrackerCommentAddRequest,
) (*domain.TrackerCommentAddResponse, error) {
	u := fmt.Sprintf("%s/v3/issues/%s/comments", c.baseURL, url.PathEscape(req.IssueID))

	isAddToFollowers := req.IsAddToFollowers
	body := AddCommentRequest{
		Text:              req.Text,
		AttachmentIDs:     req.AttachmentIDs,
		MarkupType:        req.MarkupType,
		Summonees:         req.Summonees,
		MaillistSummonees: req.MaillistSummonees,
		IsAddToFollowers:  &isAddToFollowers,
	}

	var comment Comment
	if _, err := c.doPOST(ctx, u, body, &comment, "AddComment"); err != nil {
		return nil, err
	}

	result := commentToTrackerComment(comment)
	return &domain.TrackerCommentAddResponse{Comment: result}, nil
}

// doPATCH executes a PATCH request with token injection and 401 retry logic.
func (c *Client) doPATCH(
	ctx context.Context,
	urlStr string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	resp, err := c.executeRequest(ctx, http.MethodPatch, urlStr, body, false)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		resp, err = c.executeRequest(ctx, http.MethodPatch, urlStr, body, true)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.Header, c.handleResponse(ctx, resp, result, operation)
}

// doGET executes a GET request with token injection and 401 retry logic.
func (c *Client) doGET(ctx context.Context, urlStr string, result any, operation string) (http.Header, error) {
	resp, err := c.executeRequest(ctx, http.MethodGet, urlStr, nil, false)
	if err != nil {
		return nil, err
	}

	// On 401, force token refresh and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		resp, err = c.executeRequest(ctx, http.MethodGet, urlStr, nil, true)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.Header, c.handleResponse(ctx, resp, result, operation)
}

// doPOST executes a POST request with token injection and 401 retry logic.
func (c *Client) doPOST(
	ctx context.Context,
	urlStr string,
	body any,
	result any,
	operation string,
) (http.Header, error) {
	resp, err := c.executeRequest(ctx, http.MethodPost, urlStr, body, false)
	if err != nil {
		return nil, err
	}

	// On 401, force token refresh and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		resp, err = c.executeRequest(ctx, http.MethodPost, urlStr, body, true)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.Header, c.handleResponse(ctx, resp, result, operation)
}

// executeRequest performs a single HTTP request with token injection.
func (c *Client) executeRequest(
	ctx context.Context,
	method, urlStr string,
	body any,
	forceRefresh bool,
) (*http.Response, error) {
	var token string
	var err error

	if forceRefresh {
		token, err = c.tokenProvider.ForceRefresh(ctx)
	} else {
		token, err = c.tokenProvider.Token(ctx)
	}
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("get token: %w", err))
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, errorLogWrapper(ctx, fmt.Errorf("marshal request body: %w", err))
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("create request: %w", err))
	}

	req.Header.Set(headerAuthorization, "Bearer "+token)
	req.Header.Set(headerCloudOrgID, c.orgID)
	req.Header.Set(headerAcceptLanguage, acceptLangEN)
	if body != nil {
		req.Header.Set(headerContentType, contentTypeJSON)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("execute request: %w", err))
	}
	return resp, nil
}

// handleResponse processes the HTTP response and decodes the result.
func (c *Client) handleResponse(ctx context.Context, resp *http.Response, result any, operation string) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorLogWrapper(ctx, fmt.Errorf("read response body: %w", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseError(ctx, resp.StatusCode, bodyBytes, operation)
	}

	if result != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return errorLogWrapper(ctx, fmt.Errorf("decode response: %w", err))
		}
	}

	return nil
}

// parseError converts an HTTP error response into a domain.UpstreamError.
func (c *Client) parseError(ctx context.Context, statusCode int, body []byte, operation string) error {
	var errResp errorResponse
	var message string

	// Attempt to parse structured error
	if err := json.Unmarshal(body, &errResp); err == nil {
		if len(errResp.ErrorMessages) > 0 {
			message = strings.Join(errResp.ErrorMessages, "; ")
		} else if len(errResp.Errors) > 0 {
			message = strings.Join(errResp.Errors, "; ")
		}
	}

	if message == "" {
		message = http.StatusText(statusCode)
	}

	err := domain.NewUpstreamError(
		domain.ServiceTracker,
		operation,
		statusCode,
		"",
		message,
		string(body),
	)

	return errorLogWrapper(ctx, err)
}

// parseIntHeader parses an integer from a header value.
func parseIntHeader(headers http.Header, key string) (int, bool) {
	val := headers.Get(key)
	if val == "" {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseIntHeaderValue is a convenience wrapper that returns just the value (0 if header absent or invalid).
func parseIntHeaderValue(headers http.Header, key string) int {
	v, _ := parseIntHeader(headers, key)
	return v
}

// stringMapToAnyMap converts map[string]string to map[string]any for API request bodies.
func stringMapToAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

func errorLogWrapper(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	slog.ErrorContext(ctx, "tracker adapter error", "error", err)
	return err
}
