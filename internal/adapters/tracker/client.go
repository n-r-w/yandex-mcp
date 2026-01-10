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

	reqBody := searchRequestDTO{
		Filter: apihelpers.StringMapToAnyMap(opts.Filter),
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
	u, err := url.Parse(c.baseURL + "/v3/issues/_count")
	if err != nil {
		return 0, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	reqBody := countRequestDTO{
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

// CreateIssue creates a new Tracker issue.
func (c *Client) CreateIssue(
	ctx context.Context,
	req *domain.TrackerIssueCreateRequest,
) (*domain.TrackerIssueCreateResponse, error) {
	u := c.baseURL + "/v3/issues"

	body := createIssueRequestDTO{
		Queue:         req.Queue,
		Summary:       req.Summary,
		Description:   req.Description,
		Type:          req.Type,
		Priority:      req.Priority,
		Assignee:      req.Assignee,
		Tags:          req.Tags,
		Parent:        req.Parent,
		AttachmentIDs: apihelpers.StringsToStringIDs(req.AttachmentIDs),
		Sprint:        req.Sprint,
	}

	var issue issueDTO
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

	body := updateIssueRequestDTO{
		Summary:     req.Summary,
		Description: req.Description,
		Type:        req.Type,
		Priority:    req.Priority,
		Assignee:    req.Assignee,
		Version:     req.Version,
		Project:     buildUpdateIssueProject(req.ProjectPrimary, req.ProjectSecondaryAdd),
		Sprint:      buildUpdateIssueSprint(req.Sprint),
	}

	var issue issueDTO
	if _, err := c.apiClient.DoPATCH(ctx, u, body, &issue, "UpdateIssue"); err != nil {
		return nil, err
	}

	result := issueToTrackerIssue(issue)
	return &domain.TrackerIssueUpdateResponse{Issue: result}, nil
}

func buildUpdateIssueProject(primary int, secondaryAdd []int) *updateIssueProjectDTO {
	if primary == 0 && len(secondaryAdd) == 0 {
		return nil
	}

	var secondary *updateIssueProjectSecAddDTO
	if len(secondaryAdd) > 0 {
		secondary = &updateIssueProjectSecAddDTO{Add: secondaryAdd}
	}

	return &updateIssueProjectDTO{
		Primary:   primary,
		Secondary: secondary,
	}
}

func buildUpdateIssueSprint(sprintIDs []string) []updateIssueSprintDTO {
	if len(sprintIDs) == 0 {
		return nil
	}

	result := make([]updateIssueSprintDTO, len(sprintIDs))
	for i, id := range sprintIDs {
		result[i] = updateIssueSprintDTO{ID: apihelpers.StringID(id)}
	}

	return result
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
		body = executeTransitionRequestDTO{
			Comment: req.Comment,
			Fields:  req.Fields,
		}
	}

	var transitions []transitionDTO
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
	body := addCommentRequestDTO{
		Text:              req.Text,
		AttachmentIDs:     apihelpers.StringsToStringIDs(req.AttachmentIDs),
		MarkupType:        req.MarkupType,
		Summonees:         req.Summonees,
		MaillistSummonees: req.MaillistSummonees,
		IsAddToFollowers:  &isAddToFollowers,
	}

	var comment commentDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &comment, "AddComment"); err != nil {
		return nil, err
	}

	result := commentToTrackerComment(comment)
	return &domain.TrackerCommentAddResponse{Comment: result}, nil
}

// UpdateComment updates an existing comment on an issue.
func (c *Client) UpdateComment(
	ctx context.Context,
	req *domain.TrackerCommentUpdateRequest,
) (*domain.TrackerCommentUpdateResponse, error) {
	u := fmt.Sprintf("%s/v3/issues/%s/comments/%s",
		c.baseURL, url.PathEscape(req.IssueID), url.PathEscape(req.CommentID))

	body := updateCommentRequestDTO{
		Text:              req.Text,
		AttachmentIDs:     apihelpers.StringsToStringIDs(req.AttachmentIDs),
		MarkupType:        req.MarkupType,
		Summonees:         req.Summonees,
		MaillistSummonees: req.MaillistSummonees,
	}

	var comment commentDTO
	if _, err := c.apiClient.DoPATCH(ctx, u, body, &comment, "UpdateComment"); err != nil {
		return nil, err
	}

	result := commentToTrackerComment(comment)
	return &domain.TrackerCommentUpdateResponse{Comment: result}, nil
}

// DeleteComment deletes a comment from an issue.
func (c *Client) DeleteComment(ctx context.Context, req *domain.TrackerCommentDeleteRequest) error {
	u := fmt.Sprintf("%s/v3/issues/%s/comments/%s",
		c.baseURL, url.PathEscape(req.IssueID), url.PathEscape(req.CommentID))

	if _, err := c.apiClient.DoDELETE(ctx, u, "DeleteComment"); err != nil {
		return err
	}

	return nil
}

// ListIssueAttachments lists attachments for an issue.
func (c *Client) ListIssueAttachments(ctx context.Context, issueID string) ([]domain.TrackerAttachment, error) {
	u := fmt.Sprintf("%s/v3/issues/%s/attachments", c.baseURL, url.PathEscape(issueID))

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

// DeleteAttachment deletes an attachment from an issue.
func (c *Client) DeleteAttachment(ctx context.Context, req *domain.TrackerAttachmentDeleteRequest) error {
	u := fmt.Sprintf("%s/v3/issues/%s/attachments/%s/",
		c.baseURL, url.PathEscape(req.IssueID), url.PathEscape(req.FileID))

	if _, err := c.apiClient.DoDELETE(ctx, u, "DeleteAttachment"); err != nil {
		return err
	}

	return nil
}

// GetQueue gets a queue by ID or key.
func (c *Client) GetQueue(
	ctx context.Context, queueID string, opts domain.TrackerGetQueueOpts,
) (*domain.TrackerQueueDetail, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/queues/%s", c.baseURL, url.PathEscape(queueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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

// CreateQueue creates a new queue.
func (c *Client) CreateQueue(
	ctx context.Context, req *domain.TrackerQueueCreateRequest,
) (*domain.TrackerQueueCreateResponse, error) {
	u := c.baseURL + "/v3/queues/"

	body := createQueueRequestDTO{
		Key:             req.Key,
		Name:            req.Name,
		Lead:            req.Lead,
		DefaultType:     req.DefaultType,
		DefaultPriority: req.DefaultPriority,
	}

	var queue queueDetailDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &queue, "CreateQueue"); err != nil {
		return nil, err
	}

	result := queueDetailToTrackerQueueDetail(queue)
	return &domain.TrackerQueueCreateResponse{Queue: result}, nil
}

// DeleteQueue deletes a queue.
func (c *Client) DeleteQueue(ctx context.Context, req *domain.TrackerQueueDeleteRequest) error {
	u := fmt.Sprintf("%s/v3/queues/%s", c.baseURL, url.PathEscape(req.QueueID))

	if _, err := c.apiClient.DoDELETE(ctx, u, "DeleteQueue"); err != nil {
		return err
	}

	return nil
}

// RestoreQueue restores a deleted queue.
func (c *Client) RestoreQueue(
	ctx context.Context, req *domain.TrackerQueueRestoreRequest,
) (*domain.TrackerQueueRestoreResponse, error) {
	u := fmt.Sprintf("%s/v3/queues/%s/_restore", c.baseURL, url.PathEscape(req.QueueID))

	var queue queueDetailDTO
	if _, err := c.apiClient.DoPOST(ctx, u, nil, &queue, "RestoreQueue"); err != nil {
		return nil, err
	}

	result := queueDetailToTrackerQueueDetail(queue)
	return &domain.TrackerQueueRestoreResponse{Queue: result}, nil
}

// GetCurrentUser gets the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*domain.TrackerUserDetail, error) {
	u := c.baseURL + "/v3/myself"

	var user userDetailDTO
	if _, err := c.apiClient.DoGET(ctx, u, &user, "GetCurrentUser"); err != nil {
		return nil, err
	}

	result := userDetailToTrackerUserDetail(user)
	return &result, nil
}

// ListUsers lists users with optional pagination.
func (c *Client) ListUsers(ctx context.Context, opts domain.TrackerListUsersOpts) (*domain.TrackerUsersPage, error) {
	u, err := url.Parse(c.baseURL + "/v3/users")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	u := fmt.Sprintf("%s/v3/users/%s", c.baseURL, url.PathEscape(userID))

	var user userDetailDTO
	if _, err := c.apiClient.DoGET(ctx, u, &user, "GetUser"); err != nil {
		return nil, err
	}

	result := userDetailToTrackerUserDetail(user)
	return &result, nil
}

// ListIssueLinks lists all links for an issue.
func (c *Client) ListIssueLinks(ctx context.Context, issueID string) ([]domain.TrackerLink, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s/links", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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

// CreateLink creates a link between issues.
func (c *Client) CreateLink(
	ctx context.Context, req *domain.TrackerLinkCreateRequest,
) (*domain.TrackerLinkCreateResponse, error) {
	u := fmt.Sprintf("%s/v3/issues/%s/links", c.baseURL, url.PathEscape(req.IssueID))

	body := createLinkRequestDTO{
		Relationship: req.Relationship,
		Issue:        req.TargetIssue,
	}

	var link linkDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &link, "CreateLink"); err != nil {
		return nil, err
	}

	return &domain.TrackerLinkCreateResponse{
		Link: linkToTrackerLink(link),
	}, nil
}

// DeleteLink deletes a link.
func (c *Client) DeleteLink(ctx context.Context, req *domain.TrackerLinkDeleteRequest) error {
	u := fmt.Sprintf("%s/v3/issues/%s/links/%s", c.baseURL, url.PathEscape(req.IssueID), url.PathEscape(req.LinkID))

	if _, err := c.apiClient.DoDELETE(ctx, u, "DeleteLink"); err != nil {
		return err
	}
	return nil
}

// GetIssueChangelog gets the changelog for an issue.
func (c *Client) GetIssueChangelog(
	ctx context.Context, issueID string, opts domain.TrackerGetChangelogOpts,
) ([]domain.TrackerChangelogEntry, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s/changelog", c.baseURL, url.PathEscape(issueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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

// MoveIssue moves an issue to another queue.
func (c *Client) MoveIssue(
	ctx context.Context, req *domain.TrackerIssueMoveRequest,
) (*domain.TrackerIssueMoveResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/issues/%s/_move", c.baseURL, url.PathEscape(req.IssueID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	q.Set("queue", req.Queue)
	if req.InitialStatus {
		q.Set("InitialStatus", "true")
	}
	u.RawQuery = q.Encode()

	var issue issueDTO
	if _, err := c.apiClient.DoPOST(ctx, u.String(), nil, &issue, "MoveIssue"); err != nil {
		return nil, err
	}

	return &domain.TrackerIssueMoveResponse{
		Issue: issueToTrackerIssue(issue),
	}, nil
}

// ListProjectComments lists comments for a project entity.
func (c *Client) ListProjectComments(
	ctx context.Context, projectID string, opts domain.TrackerListProjectCommentsOpts,
) ([]domain.TrackerProjectComment, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/entities/project/%s/comments", c.baseURL, url.PathEscape(projectID)))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
