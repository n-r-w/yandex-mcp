package tracker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// getIssue retrieves a Tracker issue by its ID or key.
func (r *Registrator) getIssue(ctx context.Context, input getIssueInputDTO) (*issueOutputDTO, error) {
	if input.IssueID == "" {
		return nil, errors.New("issue_id_or_key is required")
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
		return nil, errors.New("per_page must be non-negative")
	}
	if input.Page < 0 {
		return nil, errors.New("page must be non-negative")
	}
	if input.PerScroll < 0 {
		return nil, errors.New("per_scroll must be non-negative")
	}
	if input.PerScroll > maxPerScroll {
		return nil, fmt.Errorf("per_scroll must not exceed %d", maxPerScroll)
	}
	if input.ScrollTTLMillis < 0 {
		return nil, errors.New("scroll_ttl_millis must be non-negative")
	}

	opts := domain.TrackerSearchIssuesOpts{
		Filter:          input.Filter,
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
	opts := domain.TrackerCountIssuesOpts{
		Filter: input.Filter,
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
		return nil, errors.New("issue_id_or_key is required")
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
		return nil, errors.New("per_page must be non-negative")
	}
	if input.Page < 0 {
		return nil, errors.New("page must be non-negative")
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
		return nil, errors.New("issue_id_or_key is required")
	}
	if input.PerPage < 0 {
		return nil, errors.New("per_page must be non-negative")
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

// listAttachments lists attachments for an issue.
func (r *Registrator) listAttachments(
	ctx context.Context, input listAttachmentsInputDTO,
) (*attachmentsListOutputDTO, error) {
	if input.IssueID == "" {
		return nil, errors.New("issue_id_or_key is required")
	}

	attachments, err := r.adapter.ListIssueAttachments(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapAttachmentsToOutput(attachments), nil
}

// getAttachment downloads an attachment for an issue.
func (r *Registrator) getAttachment(
	ctx context.Context, input getAttachmentInputDTO,
) (*attachmentContentOutputDTO, error) {
	if input.IssueID == "" {
		return nil, errors.New("issue_id_or_key is required")
	}
	if input.AttachmentID == "" {
		return nil, errors.New("attachment_id is required")
	}
	if input.FileName == "" {
		return nil, errors.New("file_name is required")
	}
	if input.SavePath == "" {
		return nil, errors.New("save_path is required")
	}

	fullPath, savedPath, err := r.resolveSavePath(input.SavePath)
	if err != nil {
		return nil, err
	}
	if !input.Override {
		if _, err := os.Stat(fullPath); err == nil {
			return nil, errors.New("save_path already exists")
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check save_path: %w", err)
		}
	}

	content, err := r.adapter.GetIssueAttachment(ctx, input.IssueID, input.AttachmentID, input.FileName)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	if err := r.writeAttachment(fullPath, input.Override, content.Data); err != nil {
		return nil, err
	}

	return mapAttachmentContentToOutput(content, savedPath), nil
}

// getAttachmentPreview downloads an attachment thumbnail for an issue.
func (r *Registrator) getAttachmentPreview(
	ctx context.Context, input getAttachmentPreviewInputDTO,
) (*attachmentContentOutputDTO, error) {
	if input.IssueID == "" {
		return nil, errors.New("issue_id_or_key is required")
	}
	if input.AttachmentID == "" {
		return nil, errors.New("attachment_id is required")
	}
	if input.SavePath == "" {
		return nil, errors.New("save_path is required")
	}

	fullPath, savedPath, err := r.resolveSavePath(input.SavePath)
	if err != nil {
		return nil, err
	}
	if !input.Override {
		if _, err := os.Stat(fullPath); err == nil {
			return nil, errors.New("save_path already exists")
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check save_path: %w", err)
		}
	}

	content, err := r.adapter.GetIssueAttachmentPreview(ctx, input.IssueID, input.AttachmentID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	if err := r.writeAttachment(fullPath, input.Override, content.Data); err != nil {
		return nil, err
	}

	return mapAttachmentContentToOutput(content, savedPath), nil
}

// resolveSavePath validates and resolves a workspace-relative save path.
func (r *Registrator) resolveSavePath(savePath string) (string, string, error) {
	cleanPath := filepath.Clean(savePath)
	if filepath.IsAbs(cleanPath) {
		return "", "", errors.New("save_path must be relative to workspace")
	}
	if cleanPath == "." {
		return "", "", errors.New("save_path must point to a file")
	}
	if cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("save_path must be within workspace")
	}

	baseDir, err := r.workspaceRoot()
	if err != nil {
		return "", "", err
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve workspace root: %w", err)
	}

	fullPath := filepath.Clean(filepath.Join(baseDir, cleanPath))
	relativePath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve save_path: %w", err)
	}
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("save_path must be within workspace")
	}

	return fullPath, relativePath, nil
}

// writeAttachment writes attachment bytes to disk.
func (r *Registrator) writeAttachment(fullPath string, override bool, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(fullPath), attachmentDirPerm); err != nil {
		return fmt.Errorf("create attachment directory: %w", err)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if override {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	//nolint:gosec // save_path is validated and constrained to workspace
	file, err := os.OpenFile(fullPath, flags, attachmentFilePerm)
	if err != nil {
		return fmt.Errorf("open save_path: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write attachment: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close attachment: %w", err)
	}

	return nil
}

// workspaceRoot returns the base directory for workspace-relative paths.
func (r *Registrator) workspaceRoot() (string, error) {
	if r.baseDir != "" {
		return r.baseDir, nil
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get workspace root: %w", err)
	}

	return baseDir, nil
}

// getQueue gets a queue by ID or key.
func (r *Registrator) getQueue(ctx context.Context, input getQueueInputDTO) (*queueDetailOutputDTO, error) {
	if input.QueueID == "" {
		return nil, errors.New("queue_id_or_key is required")
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
		return nil, errors.New("per_page must be non-negative")
	}
	if input.Page < 0 {
		return nil, errors.New("page must be non-negative")
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
		return nil, errors.New("user_id is required")
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
		return nil, errors.New("issue_id_or_key is required")
	}

	links, err := r.adapter.ListIssueLinks(ctx, input.IssueID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	return mapLinksToOutput(links), nil
}

// getChangelog gets the changelog for an issue.
func (r *Registrator) getChangelog(ctx context.Context, input getChangelogInputDTO) (*changelogOutputDTO, error) {
	if input.IssueID == "" {
		return nil, errors.New("issue_id_or_key is required")
	}
	if input.PerPage < 0 {
		return nil, errors.New("per_page must be non-negative")
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

// listProjectComments lists comments for a project.
func (r *Registrator) listProjectComments(
	ctx context.Context, input listProjectCommentsInputDTO,
) (*projectCommentsListOutputDTO, error) {
	if input.ProjectID == "" {
		return nil, errors.New("project_id is required")
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
