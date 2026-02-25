// Package tracker provides HTTP client for Yandex Tracker API.
package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
}

// Compile-time check that Client implements the tools interface.
var _ trackertools.ITrackerAdapter = (*Client)(nil)

// NewClient creates a new Tracker API client.
func NewClient(cfg *config.Config, tokenProvider apihelpers.ITokenProvider) *Client {
	client := &Client{
		apiClient: nil, // set below
	}

	client.apiClient = apihelpers.NewAPIClient(apihelpers.APIClientConfig{
		HTTPClient:    nil, // uses default
		TokenProvider: tokenProvider,
		BaseURL:       strings.TrimSuffix(cfg.TrackerBaseURL, "/"),
		OrgID:         cfg.CloudOrgID,
		ExtraHeaders: map[string]string{
			headerAcceptLanguage: acceptLangEN,
		},
		ServiceName:         string(domain.ServiceTracker),
		ParseError:          client.parseError,
		HTTPTimeout:         cfg.HTTPTimeout,
		RawResponseMaxBytes: cfg.AttachInlineMaxBytes,
	})

	return client
}

// GetIssue retrieves an issue by its ID or key.
func (c *Client) GetIssue(
	ctx context.Context,
	issueID string,
	opts domain.TrackerGetIssueOpts,
) (*domain.TrackerIssue, error) {
	u, err := url.Parse("/v3/issues/" + url.PathEscape(issueID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	if opts.Expand != "" {
		q := u.Query()
		q.Set("expand", opts.Expand)
		u.RawQuery = q.Encode()
	}

	var issue issueDTO
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
	u, err := url.Parse("/v3/issues/_search")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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

	reqBody := searchRequestDTO{
		Filter: opts.Filter,
		Query:  opts.Query,
		Order:  opts.Order,
	}

	var issues []issueDTO
	headers, err := c.apiClient.DoPOST(ctx, u.String(), reqBody, &issues, "SearchIssues")
	if err != nil {
		return nil, err
	}

	result := searchIssuesResultToTrackerIssuesPage(
		issues,
		parseIntHeaderValue(headers, headerXTotalCount),
		parseIntHeaderValue(headers, headerXTotalPages),
		headers.Get(headerXScrollID),
		headers.Get(headerXScrollToken),
		headers.Get(headerLink),
	)
	return &result, nil
}

// CountIssues counts issues matching the filter or query.
func (c *Client) CountIssues(ctx context.Context, opts domain.TrackerCountIssuesOpts) (int, error) {
	u, err := url.Parse("/v3/issues/_count")
	if err != nil {
		return 0, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	reqBody := countRequestDTO{
		Filter: opts.Filter,
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
	u, err := url.Parse(fmt.Sprintf("/v3/issues/%s/transitions", url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	var transitions []transitionDTO
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
	u, err := url.Parse("/v3/queues/")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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

	var queues []queueDTO
	headers, err := c.apiClient.DoGET(ctx, u.String(), &queues, "ListQueues")
	if err != nil {
		return nil, err
	}

	result := listQueuesResultToTrackerQueuesPage(
		queues,
		parseIntHeaderValue(headers, headerXTotalCount),
		parseIntHeaderValue(headers, headerXTotalPages),
	)
	return &result, nil
}

// ListIssueComments lists comments for an issue.
func (c *Client) ListIssueComments(
	ctx context.Context,
	issueID string,
	opts domain.TrackerListCommentsOpts,
) (*domain.TrackerCommentsPage, error) {
	u, err := url.Parse(fmt.Sprintf("/v3/issues/%s/comments", url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	q := u.Query()
	if opts.Expand != "" {
		q.Set("expand", opts.Expand)
	}
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.ID != "" {
		q.Set("id", opts.ID)
	}
	u.RawQuery = q.Encode()

	var comments []commentDTO
	headers, err := c.apiClient.DoGET(ctx, u.String(), &comments, "ListIssueComments")
	if err != nil {
		return nil, err
	}

	result := listCommentsResultToTrackerCommentsPage(
		comments,
		headers.Get(headerLink),
	)
	return &result, nil
}

// ListIssueAttachments lists attachments for an issue.
func (c *Client) ListIssueAttachments(ctx context.Context, issueID string) ([]domain.TrackerAttachment, error) {
	u := fmt.Sprintf("/v3/issues/%s/attachments", url.PathEscape(issueID))

	var attachments []attachmentDTO
	if _, err := c.apiClient.DoGET(ctx, u, &attachments, "ListIssueAttachments"); err != nil {
		return nil, err
	}

	result := make([]domain.TrackerAttachment, len(attachments))
	for i, a := range attachments {
		result[i] = attachmentToTrackerAttachment(a)
	}

	return result, nil
}

// GetIssueAttachment downloads an attachment for an issue.
func (c *Client) GetIssueAttachment(
	ctx context.Context,
	issueID string,
	attachmentID string,
	fileName string,
) (*domain.TrackerAttachmentContent, error) {
	u := fmt.Sprintf(
		"/v3/issues/%s/attachments/%s/%s",
		url.PathEscape(issueID),
		url.PathEscape(attachmentID),
		url.PathEscape(fileName),
	)

	headers, body, err := c.apiClient.DoGETRaw(ctx, u, "GetIssueAttachment")
	if err != nil {
		return nil, err
	}

	return &domain.TrackerAttachmentContent{
		FileName:    fileName,
		ContentType: headers.Get(apihelpers.HeaderContentType),
		Data:        body,
	}, nil
}

// attachmentStream wraps an io.ReadCloser for domain streaming.
type attachmentStream struct {
	reader io.ReadCloser
}

var _ io.ReadCloser = (*attachmentStream)(nil)

// Read reads data from the underlying stream.
func (s *attachmentStream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

// Close closes the underlying stream.
func (s *attachmentStream) Close() error {
	return s.reader.Close()
}

// Compile-time check that attachmentStream implements domain stream interface.
var _ domain.IAttachmentStream = (*attachmentStream)(nil)

// GetIssueAttachmentStream streams an attachment for an issue.
func (c *Client) GetIssueAttachmentStream(
	ctx context.Context,
	issueID string,
	attachmentID string,
	fileName string,
) (*domain.TrackerAttachmentStream, error) {
	u := fmt.Sprintf(
		"/v3/issues/%s/attachments/%s/%s",
		url.PathEscape(issueID),
		url.PathEscape(attachmentID),
		url.PathEscape(fileName),
	)

	headers, body, err := c.apiClient.DoGETStream(ctx, u, "GetIssueAttachment")
	if err != nil {
		return nil, err
	}

	return &domain.TrackerAttachmentStream{
		FileName:    fileName,
		ContentType: headers.Get(apihelpers.HeaderContentType),
		Stream:      &attachmentStream{reader: body},
	}, nil
}

// GetIssueAttachmentPreview downloads an attachment thumbnail for an issue.
func (c *Client) GetIssueAttachmentPreview(
	ctx context.Context,
	issueID string,
	attachmentID string,
) (*domain.TrackerAttachmentContent, error) {
	u := fmt.Sprintf(
		"/v3/issues/%s/thumbnails/%s",
		url.PathEscape(issueID),
		url.PathEscape(attachmentID),
	)

	headers, body, err := c.apiClient.DoGETRaw(ctx, u, "GetIssueAttachmentPreview")
	if err != nil {
		return nil, err
	}

	return &domain.TrackerAttachmentContent{
		FileName:    "",
		ContentType: headers.Get(apihelpers.HeaderContentType),
		Data:        body,
	}, nil
}

// GetIssueAttachmentPreviewStream streams an attachment thumbnail for an issue.
func (c *Client) GetIssueAttachmentPreviewStream(
	ctx context.Context,
	issueID string,
	attachmentID string,
) (*domain.TrackerAttachmentStream, error) {
	u := fmt.Sprintf(
		"/v3/issues/%s/thumbnails/%s",
		url.PathEscape(issueID),
		url.PathEscape(attachmentID),
	)

	headers, body, err := c.apiClient.DoGETStream(ctx, u, "GetIssueAttachmentPreview")
	if err != nil {
		return nil, err
	}

	return &domain.TrackerAttachmentStream{
		FileName:    "",
		ContentType: headers.Get(apihelpers.HeaderContentType),
		Stream:      &attachmentStream{reader: body},
	}, nil
}

// GetQueue gets a queue by ID or key.
func (c *Client) GetQueue(
	ctx context.Context, queueID string, opts domain.TrackerGetQueueOpts,
) (*domain.TrackerQueueDetail, error) {
	u, err := url.Parse("/v3/queues/" + url.PathEscape(queueID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	if opts.Expand != "" {
		q := u.Query()
		q.Set("expand", opts.Expand)
		u.RawQuery = q.Encode()
	}

	var queue queueDetailDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &queue, "GetQueue"); err != nil {
		return nil, err
	}

	result := queueDetailToTrackerQueueDetail(queue)
	return &result, nil
}

// GetCurrentUser gets the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*domain.TrackerUserDetail, error) {
	u := "/v3/myself"

	var user userDetailDTO
	if _, err := c.apiClient.DoGET(ctx, u, &user, "GetCurrentUser"); err != nil {
		return nil, err
	}

	result := userDetailToTrackerUserDetail(user)
	return &result, nil
}

// ListUsers lists users with optional pagination.
func (c *Client) ListUsers(ctx context.Context, opts domain.TrackerListUsersOpts) (*domain.TrackerUsersPage, error) {
	u, err := url.Parse("/v3/users")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	q := u.Query()
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	u.RawQuery = q.Encode()

	var users []userDetailDTO
	headers, err := c.apiClient.DoGET(ctx, u.String(), &users, "ListUsers")
	if err != nil {
		return nil, err
	}

	result := make([]domain.TrackerUserDetail, len(users))
	for i, user := range users {
		result[i] = userDetailToTrackerUserDetail(user)
	}

	return &domain.TrackerUsersPage{
		Users:      result,
		TotalCount: parseIntHeaderValue(headers, headerXTotalCount),
		TotalPages: parseIntHeaderValue(headers, headerXTotalPages),
	}, nil
}

// GetUser gets a user by ID or login.
func (c *Client) GetUser(ctx context.Context, userID string) (*domain.TrackerUserDetail, error) {
	u := "/v3/users/" + url.PathEscape(userID)

	var user userDetailDTO
	if _, err := c.apiClient.DoGET(ctx, u, &user, "GetUser"); err != nil {
		return nil, err
	}

	result := userDetailToTrackerUserDetail(user)
	return &result, nil
}

// ListIssueLinks lists all links for an issue.
func (c *Client) ListIssueLinks(ctx context.Context, issueID string) ([]domain.TrackerLink, error) {
	u, err := url.Parse(fmt.Sprintf("/v3/issues/%s/links", url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	var links []linkDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &links, "ListIssueLinks"); err != nil {
		return nil, err
	}

	result := make([]domain.TrackerLink, len(links))
	for i, link := range links {
		result[i] = linkToTrackerLink(link)
	}
	return result, nil
}

// GetIssueChangelog gets the changelog for an issue.
func (c *Client) GetIssueChangelog(
	ctx context.Context, issueID string, opts domain.TrackerGetChangelogOpts,
) ([]domain.TrackerChangelogEntry, error) {
	u, err := url.Parse(fmt.Sprintf("/v3/issues/%s/changelog", url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	if opts.PerPage > 0 {
		q := u.Query()
		q.Set("perPage", strconv.Itoa(opts.PerPage))
		u.RawQuery = q.Encode()
	}

	var entries []changelogEntryDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &entries, "GetIssueChangelog"); err != nil {
		return nil, err
	}

	result := make([]domain.TrackerChangelogEntry, len(entries))
	for i, entry := range entries {
		result[i] = changelogEntryToTrackerChangelogEntry(entry)
	}
	return result, nil
}

// ListProjectComments lists comments for a project entity.
func (c *Client) ListProjectComments(
	ctx context.Context, projectID string, opts domain.TrackerListProjectCommentsOpts,
) ([]domain.TrackerProjectComment, error) {
	u, err := url.Parse(fmt.Sprintf("/v3/entities/project/%s/comments", url.PathEscape(projectID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
	}

	if opts.Expand != "" {
		q := u.Query()
		q.Set("expand", opts.Expand)
		u.RawQuery = q.Encode()
	}

	var comments []projectCommentDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &comments, "ListProjectComments"); err != nil {
		return nil, err
	}

	result := make([]domain.TrackerProjectComment, len(comments))
	for i, comment := range comments {
		result[i] = projectCommentToTrackerProjectComment(comment)
	}
	return result, nil
}

// parseError converts an HTTP error response into a domain.UpstreamError.
func (c *Client) parseError(ctx context.Context, statusCode int, body []byte, operation string) error {
	var errResp errorResponseDTO
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
