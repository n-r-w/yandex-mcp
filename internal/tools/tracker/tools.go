package tracker

import (
	"context"
	"errors"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// GetIssue retrieves a Tracker issue by its ID or key.
func (r *Registrator) GetIssue(ctx context.Context, input GetIssueInput) (*IssueOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	opts := domain.TrackerGetIssueOpts{
		Expand: input.Expand,
	}

	issue, err := r.adapter.GetIssue(ctx, input.IssueID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapIssueToOutput(issue), nil
}

// SearchIssues searches for Tracker issues using filter or query.
func (r *Registrator) SearchIssues(ctx context.Context, input SearchIssuesInput) (*SearchIssuesOutput, error) {
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
	}
	if input.Page < 0 {
		return nil, r.logError(ctx, errors.New("page must be non-negative"))
	}
	if input.PerScroll < 0 {
		return nil, r.logError(ctx, errors.New("per_scroll must be non-negative"))
	}
	if input.PerScroll > maxPerScroll {
		return nil, r.logError(ctx, fmt.Errorf("per_scroll must not exceed %d", maxPerScroll))
	}
	if input.ScrollTTLMillis < 0 {
		return nil, r.logError(ctx, errors.New("scroll_ttl_millis must be non-negative"))
	}

	filter, err := helpers.ConvertFilterToStringMap(ctx, input.Filter, domain.ServiceTracker)
	if err != nil {
		return nil, err
	}

	opts := domain.TrackerSearchIssuesOpts{
		Filter:          filter,
		Query:           input.Query,
		Order:           input.Order,
		Expand:          input.Expand,
		PerPage:         input.PerPage,
		Page:            input.Page,
		ScrollType:      input.ScrollType,
		PerScroll:       input.PerScroll,
		ScrollTTLMillis: input.ScrollTTLMillis,
		ScrollID:        input.ScrollID,
	}

	result, err := r.adapter.SearchIssues(ctx, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapSearchResultToOutput(result), nil
}

// CountIssues counts Tracker issues matching the filter or query.
func (r *Registrator) CountIssues(ctx context.Context, input CountIssuesInput) (*CountIssuesOutput, error) {
	filter, err := helpers.ConvertFilterToStringMap(ctx, input.Filter, domain.ServiceTracker)
	if err != nil {
		return nil, err
	}

	opts := domain.TrackerCountIssuesOpts{
		Filter: filter,
		Query:  input.Query,
	}

	count, err := r.adapter.CountIssues(ctx, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return &CountIssuesOutput{Count: count}, nil
}

// ListTransitions lists available transitions for a Tracker issue.
func (r *Registrator) ListTransitions(ctx context.Context, input ListTransitionsInput) (*TransitionsListOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	transitions, err := r.adapter.ListIssueTransitions(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapTransitionsToOutput(transitions), nil
}

// ListQueues lists all Tracker queues.
func (r *Registrator) ListQueues(ctx context.Context, input ListQueuesInput) (*QueuesListOutput, error) {
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
	}
	if input.Page < 0 {
		return nil, r.logError(ctx, errors.New("page must be non-negative"))
	}

	opts := domain.TrackerListQueuesOpts{
		Expand:  input.Expand,
		PerPage: input.PerPage,
		Page:    input.Page,
	}

	result, err := r.adapter.ListQueues(ctx, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapQueuesResultToOutput(result), nil
}

// ListComments lists comments for a Tracker issue.
func (r *Registrator) ListComments(ctx context.Context, input ListCommentsInput) (*CommentsListOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
	}
	if input.ID < 0 {
		return nil, r.logError(ctx, errors.New("id must be non-negative"))
	}

	opts := domain.TrackerListCommentsOpts{
		Expand:  input.Expand,
		PerPage: input.PerPage,
		ID:      input.ID,
	}

	result, err := r.adapter.ListIssueComments(ctx, input.IssueID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapCommentsResultToOutput(result), nil
}

// CreateIssue creates a new Tracker issue.
func (r *Registrator) CreateIssue(ctx context.Context, input CreateIssueInput) (*IssueOutput, error) {
	if input.Queue == "" {
		return nil, r.logError(ctx, errors.New("queue is required"))
	}

	if input.Summary == "" {
		return nil, r.logError(ctx, errors.New("summary is required"))
	}

	req := &domain.TrackerIssueCreateRequest{
		Queue:         input.Queue,
		Summary:       input.Summary,
		Description:   input.Description,
		Type:          input.Type,
		Priority:      input.Priority,
		Assignee:      input.Assignee,
		Tags:          input.Tags,
		Parent:        input.Parent,
		AttachmentIDs: input.AttachmentIDs,
		Sprint:        input.Sprint,
	}

	result, err := r.adapter.CreateIssue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapIssueToOutput(&result.Issue), nil
}

// UpdateIssue updates an existing Tracker issue.
func (r *Registrator) UpdateIssue(ctx context.Context, input UpdateIssueInput) (*IssueOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.Summary == "" && input.Description == "" && input.Type == "" && input.Priority == "" && input.Assignee == "" {
		return nil, r.logError(ctx, errors.New("at least one field to update is required"))
	}

	req := &domain.TrackerIssueUpdateRequest{
		IssueID:     input.IssueID,
		Summary:     input.Summary,
		Description: input.Description,
		Type:        input.Type,
		Priority:    input.Priority,
		Assignee:    input.Assignee,
		Version:     input.Version,
	}

	result, err := r.adapter.UpdateIssue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapIssueToOutput(&result.Issue), nil
}

// ExecuteTransition executes a status transition on an issue.
func (r *Registrator) ExecuteTransition(
	ctx context.Context, input ExecuteTransitionInput,
) (*TransitionsListOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.TransitionID == "" {
		return nil, r.logError(ctx, errors.New("transition_id is required"))
	}

	req := &domain.TrackerTransitionExecuteRequest{
		IssueID:      input.IssueID,
		TransitionID: input.TransitionID,
		Comment:      input.Comment,
		Fields:       input.Fields,
	}

	result, err := r.adapter.ExecuteTransition(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapTransitionsToOutput(result.Transitions), nil
}

// AddComment adds a comment to an issue.
func (r *Registrator) AddComment(ctx context.Context, input AddCommentInput) (*CommentOutput, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.Text == "" {
		return nil, r.logError(ctx, errors.New("text is required"))
	}

	req := &domain.TrackerCommentAddRequest{
		IssueID:           input.IssueID,
		Text:              input.Text,
		AttachmentIDs:     input.AttachmentIDs,
		MarkupType:        input.MarkupType,
		Summonees:         input.Summonees,
		MaillistSummonees: input.MaillistSummonees,
		IsAddToFollowers:  input.IsAddToFollowers,
	}

	result, err := r.adapter.AddComment(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapCommentToOutput(&result.Comment), nil
}
