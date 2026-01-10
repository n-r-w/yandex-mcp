package tracker

import (
	"context"
	"errors"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// getIssue retrieves a Tracker issue by its ID or key.
func (r *Registrator) getIssue(ctx context.Context, input getIssueInputDTO) (*issueOutputDTO, error) {
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

// searchIssues searches for Tracker issues using filter or query.
func (r *Registrator) searchIssues(ctx context.Context, input searchIssuesInputDTO) (*searchIssuesOutputDTO, error) {
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

// countIssues counts Tracker issues matching the filter or query.
func (r *Registrator) countIssues(ctx context.Context, input countIssuesInputDTO) (*countIssuesOutputDTO, error) {
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

	return &countIssuesOutputDTO{Count: count}, nil
}

// listTransitions lists available transitions for a Tracker issue.
func (r *Registrator) listTransitions(
	ctx context.Context, input listTransitionsInputDTO,
) (*transitionsListOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	transitions, err := r.adapter.ListIssueTransitions(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapTransitionsToOutput(transitions), nil
}

// listQueues lists all Tracker queues.
func (r *Registrator) listQueues(ctx context.Context, input listQueuesInputDTO) (*queuesListOutputDTO, error) {
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

// listComments lists comments for a Tracker issue.
func (r *Registrator) listComments(ctx context.Context, input listCommentsInputDTO) (*commentsListOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
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

// createIssue creates a new Tracker issue.
func (r *Registrator) createIssue(ctx context.Context, input createIssueInputDTO) (*issueOutputDTO, error) {
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

// updateIssue updates an existing Tracker issue.
func (r *Registrator) updateIssue(ctx context.Context, input updateIssueInputDTO) (*issueOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	hasBasicFields := input.Summary != "" || input.Description != "" ||
		input.Type != "" || input.Priority != "" || input.Assignee != ""
	hasProjectFields := input.ProjectPrimary != 0 || len(input.ProjectSecondaryAdd) > 0
	hasSprintFields := len(input.Sprint) > 0

	if !hasBasicFields && !hasProjectFields && !hasSprintFields {
		return nil, r.logError(ctx, errors.New("at least one field to update is required"))
	}

	req := &domain.TrackerIssueUpdateRequest{
		IssueID:             input.IssueID,
		Summary:             input.Summary,
		Description:         input.Description,
		Type:                input.Type,
		Priority:            input.Priority,
		Assignee:            input.Assignee,
		Version:             input.Version,
		ProjectPrimary:      input.ProjectPrimary,
		ProjectSecondaryAdd: input.ProjectSecondaryAdd,
		Sprint:              input.Sprint,
	}

	result, err := r.adapter.UpdateIssue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapIssueToOutput(&result.Issue), nil
}

// executeTransition executes a status transition on an issue.
func (r *Registrator) executeTransition(
	ctx context.Context, input executeTransitionInputDTO,
) (*transitionsListOutputDTO, error) {
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

// addComment adds a comment to an issue.
func (r *Registrator) addComment(ctx context.Context, input addCommentInputDTO) (*commentOutputDTO, error) {
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

// updateComment updates an existing comment on an issue.
func (r *Registrator) updateComment(ctx context.Context, input updateCommentInputDTO) (*commentOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.CommentID == "" {
		return nil, r.logError(ctx, errors.New("comment_id is required"))
	}

	if input.Text == "" {
		return nil, r.logError(ctx, errors.New("text is required"))
	}

	req := &domain.TrackerCommentUpdateRequest{
		IssueID:           input.IssueID,
		CommentID:         input.CommentID,
		Text:              input.Text,
		AttachmentIDs:     input.AttachmentIDs,
		MarkupType:        input.MarkupType,
		Summonees:         input.Summonees,
		MaillistSummonees: input.MaillistSummonees,
	}

	result, err := r.adapter.UpdateComment(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapCommentToOutput(&result.Comment), nil
}

// deleteComment deletes a comment from an issue.
func (r *Registrator) deleteComment(ctx context.Context, input deleteCommentInputDTO) (*deleteCommentOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.CommentID == "" {
		return nil, r.logError(ctx, errors.New("comment_id is required"))
	}

	req := &domain.TrackerCommentDeleteRequest{
		IssueID:   input.IssueID,
		CommentID: input.CommentID,
	}

	if err := r.adapter.DeleteComment(ctx, req); err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return &deleteCommentOutputDTO{Success: true}, nil
}

// listAttachments lists attachments for an issue.
func (r *Registrator) listAttachments(
	ctx context.Context, input listAttachmentsInputDTO,
) (*attachmentsListOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	attachments, err := r.adapter.ListIssueAttachments(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapAttachmentsToOutput(attachments), nil
}

// deleteAttachment deletes an attachment from an issue.
func (r *Registrator) deleteAttachment(
	ctx context.Context, input deleteAttachmentInputDTO,
) (*deleteAttachmentOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	if input.FileID == "" {
		return nil, r.logError(ctx, errors.New("file_id is required"))
	}

	req := &domain.TrackerAttachmentDeleteRequest{
		IssueID: input.IssueID,
		FileID:  input.FileID,
	}

	if err := r.adapter.DeleteAttachment(ctx, req); err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return &deleteAttachmentOutputDTO{Success: true}, nil
}

// getQueue gets a queue by ID or key.
func (r *Registrator) getQueue(ctx context.Context, input getQueueInputDTO) (*queueDetailOutputDTO, error) {
	if input.QueueID == "" {
		return nil, r.logError(ctx, errors.New("queue_id_or_key is required"))
	}

	opts := domain.TrackerGetQueueOpts{
		Expand: input.Expand,
	}

	queue, err := r.adapter.GetQueue(ctx, input.QueueID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapQueueDetailToOutput(queue), nil
}

// createQueue creates a new queue.
func (r *Registrator) createQueue(ctx context.Context, input createQueueInputDTO) (*queueDetailOutputDTO, error) {
	if input.Key == "" {
		return nil, r.logError(ctx, errors.New("key is required"))
	}
	if input.Name == "" {
		return nil, r.logError(ctx, errors.New("name is required"))
	}
	if input.Lead == "" {
		return nil, r.logError(ctx, errors.New("lead is required"))
	}
	if input.DefaultType == "" {
		return nil, r.logError(ctx, errors.New("default_type is required"))
	}
	if input.DefaultPriority == "" {
		return nil, r.logError(ctx, errors.New("default_priority is required"))
	}

	req := &domain.TrackerQueueCreateRequest{
		Key:             input.Key,
		Name:            input.Name,
		Lead:            input.Lead,
		DefaultType:     input.DefaultType,
		DefaultPriority: input.DefaultPriority,
	}

	resp, err := r.adapter.CreateQueue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapQueueDetailToOutput(&resp.Queue), nil
}

// deleteQueue deletes a queue.
func (r *Registrator) deleteQueue(ctx context.Context, input deleteQueueInputDTO) (*deleteQueueOutputDTO, error) {
	if input.QueueID == "" {
		return nil, r.logError(ctx, errors.New("queue_id_or_key is required"))
	}

	req := &domain.TrackerQueueDeleteRequest{
		QueueID: input.QueueID,
	}

	if err := r.adapter.DeleteQueue(ctx, req); err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return &deleteQueueOutputDTO{Success: true}, nil
}

// restoreQueue restores a deleted queue.
func (r *Registrator) restoreQueue(ctx context.Context, input restoreQueueInputDTO) (*queueDetailOutputDTO, error) {
	if input.QueueID == "" {
		return nil, r.logError(ctx, errors.New("queue_id_or_key is required"))
	}

	req := &domain.TrackerQueueRestoreRequest{
		QueueID: input.QueueID,
	}

	resp, err := r.adapter.RestoreQueue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapQueueDetailToOutput(&resp.Queue), nil
}

// getCurrentUser gets the current authenticated user.
func (r *Registrator) getCurrentUser(ctx context.Context, _ getCurrentUserInputDTO) (*userDetailOutputDTO, error) {
	user, err := r.adapter.GetCurrentUser(ctx)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapUserDetailToOutput(user), nil
}

// listUsers lists users with optional pagination.
func (r *Registrator) listUsers(ctx context.Context, input listUsersInputDTO) (*usersListOutputDTO, error) {
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
	}
	if input.Page < 0 {
		return nil, r.logError(ctx, errors.New("page must be non-negative"))
	}

	opts := domain.TrackerListUsersOpts{
		PerPage: input.PerPage,
		Page:    input.Page,
	}

	result, err := r.adapter.ListUsers(ctx, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapUsersPageToOutput(result), nil
}

// getUser gets a user by ID or login.
func (r *Registrator) getUser(ctx context.Context, input getUserInputDTO) (*userDetailOutputDTO, error) {
	if input.UserID == "" {
		return nil, r.logError(ctx, errors.New("user_id is required"))
	}

	user, err := r.adapter.GetUser(ctx, input.UserID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapUserDetailToOutput(user), nil
}

// listLinks lists all links for an issue.
func (r *Registrator) listLinks(ctx context.Context, input listLinksInputDTO) (*linksListOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}

	links, err := r.adapter.ListIssueLinks(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapLinksToOutput(links), nil
}

// createLink creates a link between issues.
func (r *Registrator) createLink(ctx context.Context, input createLinkInputDTO) (*linkOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.Relationship == "" {
		return nil, r.logError(ctx, errors.New("relationship is required"))
	}
	if input.TargetIssue == "" {
		return nil, r.logError(ctx, errors.New("target_issue is required"))
	}

	req := &domain.TrackerLinkCreateRequest{
		IssueID:      input.IssueID,
		Relationship: input.Relationship,
		TargetIssue:  input.TargetIssue,
	}

	result, err := r.adapter.CreateLink(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	out := mapLinkToOutput(&result.Link)
	return &out, nil
}

// deleteLink deletes a link.
func (r *Registrator) deleteLink(ctx context.Context, input deleteLinkInputDTO) (*deleteLinkOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.LinkID == "" {
		return nil, r.logError(ctx, errors.New("link_id is required"))
	}

	req := &domain.TrackerLinkDeleteRequest{
		IssueID: input.IssueID,
		LinkID:  input.LinkID,
	}

	if err := r.adapter.DeleteLink(ctx, req); err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return &deleteLinkOutputDTO{Success: true}, nil
}

// getChangelog gets the changelog for an issue.
func (r *Registrator) getChangelog(ctx context.Context, input getChangelogInputDTO) (*changelogOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.PerPage < 0 {
		return nil, r.logError(ctx, errors.New("per_page must be non-negative"))
	}

	opts := domain.TrackerGetChangelogOpts{
		PerPage: input.PerPage,
	}

	entries, err := r.adapter.GetIssueChangelog(ctx, input.IssueID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapChangelogToOutput(entries), nil
}

// moveIssue moves an issue to another queue.
func (r *Registrator) moveIssue(ctx context.Context, input moveIssueInputDTO) (*issueOutputDTO, error) {
	if input.IssueID == "" {
		return nil, r.logError(ctx, errors.New("issue_id_or_key is required"))
	}
	if input.Queue == "" {
		return nil, r.logError(ctx, errors.New("queue is required"))
	}

	req := &domain.TrackerIssueMoveRequest{
		IssueID:       input.IssueID,
		Queue:         input.Queue,
		InitialStatus: input.InitialStatus,
	}

	result, err := r.adapter.MoveIssue(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapIssueToOutput(&result.Issue), nil
}

// listProjectComments lists comments for a project.
func (r *Registrator) listProjectComments(
	ctx context.Context, input listProjectCommentsInputDTO,
) (*projectCommentsListOutputDTO, error) {
	if input.ProjectID == "" {
		return nil, r.logError(ctx, errors.New("project_id is required"))
	}

	opts := domain.TrackerListProjectCommentsOpts{
		Expand: input.Expand,
	}

	comments, err := r.adapter.ListProjectComments(ctx, input.ProjectID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapProjectCommentsToOutput(comments), nil
}
