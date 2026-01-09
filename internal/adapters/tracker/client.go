// Package tracker provides HTTP client for Yandex Tracker API.
package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	trackertools "github.com/n-r-w/yandex-mcp/internal/tools/tracker"
)

// Client implements HTTP client for Yandex Tracker API.
type Client struct {
	apiClient *apihelpers.APIClient
	baseURL   string
}

// Compile-time check that Client implements the tools interface.
var _ trackertools.ITrackerAdapter = (*Client)(nil)

// NewClient creates a new Tracker API client.
func NewClient(cfg *config.Config, tokenProvider apihelpers.ITokenProvider) *Client {
	client := &Client{
		apiClient: nil, // set below
		baseURL:   strings.TrimSuffix(cfg.TrackerBaseURL, "/"),
	}

	client.apiClient = apihelpers.NewAPIClient(apihelpers.APIClientConfig{
		HTTPClient:    nil, // uses default
		TokenProvider: tokenProvider,
		OrgID:         cfg.CloudOrgID,
		ExtraHeaders: map[string]string{
			headerAcceptLanguage: acceptLangEN,
		},
		ServiceName: string(domain.ServiceTracker),
		ParseError:  client.parseError,
		HTTPTimeout: cfg.HTTPTimeout,
	})

	return client
}

// GetIssue retrieves an issue by its ID or key.
func (c *Client) GetIssue(
	ctx context.Context,
	issueID string,
	opts domain.TrackerGetIssueOpts,
) (*domain.TrackerIssue, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	if opts.Expand != "" {
		q := u.Query()
		q.Set("expand", opts.Expand)
		u.RawQuery = q.Encode()
	}

	var issue Issue
	if _, err := c.apiClient.DoGET(ctx, u.String(), &issue, "GetIssue"); err != nil {
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
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
		Filter: apihelpers.StringMapToAnyMap(opts.Filter),
		Query:  opts.Query,
		Order:  opts.Order,
	}

	var issues []Issue
	headers, err := c.apiClient.DoPOST(ctx, u.String(), reqBody, &issues, "SearchIssues")
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
		return 0, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	reqBody := countRequest{
		Filter: apihelpers.StringMapToAnyMap(opts.Filter),
		Query:  opts.Query,
	}

	var count int
	if _, err := c.apiClient.DoPOST(ctx, u.String(), reqBody, &count, "CountIssues"); err != nil {
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
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	var transitions []Transition
	if _, err := c.apiClient.DoGET(ctx, u.String(), &transitions, "ListIssueTransitions"); err != nil {
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
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	headers, err := c.apiClient.DoGET(ctx, u.String(), &queues, "ListQueues")
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
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	headers, err := c.apiClient.DoGET(ctx, u.String(), &comments, "ListIssueComments")
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
	if _, err := c.apiClient.DoPOST(ctx, u, body, &issue, "CreateIssue"); err != nil {
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
	if _, err := c.apiClient.DoPATCH(ctx, u, body, &issue, "UpdateIssue"); err != nil {
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
	if _, err := c.apiClient.DoPOST(ctx, u, body, &transitions, "ExecuteTransition"); err != nil {
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
	if _, err := c.apiClient.DoPOST(ctx, u, body, &comment, "AddComment"); err != nil {
		return nil, err
	}

	result := commentToTrackerComment(comment)
	return &domain.TrackerCommentAddResponse{Comment: result}, nil
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

	return c.apiClient.ErrorLogWrapper(ctx, err)
}

// parseIntHeaderValue parses an integer from a header value, returning 0 if absent or invalid.
func parseIntHeaderValue(headers http.Header, key string) int {
	val := headers.Get(key)
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
}
