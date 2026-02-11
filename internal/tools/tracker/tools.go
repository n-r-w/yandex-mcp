package tracker

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	if input.SavePath == "" && !input.GetContent {
		return nil, errors.New("save_path or get_content is required")
	}
	if input.SavePath != "" && input.GetContent {
		return nil, errors.New("save_path and get_content cannot be used together")
	}
	if input.GetContent {
		if err := r.validateAttachmentViewExtension(input.FileName); err != nil {
			return nil, err
		}
	}

	var (
		fullPath  string
		savedPath string
	)
	if input.SavePath != "" {
		var err error
		fullPath, savedPath, err = r.prepareSavePath(ctx, input.SavePath, input.Override)
		if err != nil {
			return nil, err
		}
	}
	if input.SavePath != "" {
		stream, err := r.adapter.GetIssueAttachmentStream(ctx, input.IssueID, input.AttachmentID, input.FileName)
		if err != nil {
			return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
		}
		if stream == nil || stream.Stream == nil {
			return nil, r.logError(ctx, errors.New("attachment stream is empty"))
		}
		bytesWritten, writeErr := r.writeAttachmentStream(fullPath, input.Override, stream.Stream)
		closeErr := stream.Stream.Close()
		if writeErr != nil {
			return nil, r.logError(ctx, writeErr)
		}
		if closeErr != nil {
			return nil, r.logError(ctx, fmt.Errorf("close attachment stream: %w", closeErr))
		}
		return &attachmentContentOutputDTO{
			FileName:    stream.FileName,
			ContentType: stream.ContentType,
			SavedPath:   savedPath,
			Content:     "",
			Size:        bytesWritten,
		}, nil
	}

	content, err := r.adapter.GetIssueAttachment(ctx, input.IssueID, input.AttachmentID, input.FileName)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}

	inlineContent := ""
	if input.GetContent {
		inlineContent = string(content.Data)
	}
	return mapAttachmentContentToOutput(content, savedPath, inlineContent), nil
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

	fullPath, savedPath, err := r.prepareSavePath(ctx, input.SavePath, input.Override)
	if err != nil {
		return nil, err
	}
	stream, err := r.adapter.GetIssueAttachmentPreviewStream(ctx, input.IssueID, input.AttachmentID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceTracker, err)
	}
	if stream == nil || stream.Stream == nil {
		return nil, r.logError(ctx, errors.New("attachment stream is empty"))
	}
	bytesWritten, writeErr := r.writeAttachmentStream(fullPath, input.Override, stream.Stream)
	closeErr := stream.Stream.Close()
	if writeErr != nil {
		return nil, r.logError(ctx, writeErr)
	}
	if closeErr != nil {
		return nil, r.logError(ctx, fmt.Errorf("close attachment stream: %w", closeErr))
	}

	return &attachmentContentOutputDTO{
		FileName:    stream.FileName,
		ContentType: stream.ContentType,
		SavedPath:   savedPath,
		Content:     "",
		Size:        bytesWritten,
	}, nil
}

// resolveSavePath validates the absolute save path against attachment safety rules.
func (r *Registrator) resolveSavePath(ctx context.Context, savePath string) (string, string, error) {
	cleanPath := filepath.Clean(savePath)
	if !filepath.IsAbs(cleanPath) {
		return "", "", fmt.Errorf("save_path must be absolute; allowed paths: %s", r.allowedPathsSummary())
	}
	if cleanPath == "." || cleanPath == string(os.PathSeparator) {
		return "", "", fmt.Errorf("save_path must point to a file; allowed paths: %s", r.allowedPathsSummary())
	}
	if err := r.validateAttachmentExtension(cleanPath); err != nil {
		return "", "", err
	}
	if err := r.validateAttachmentDirectory(ctx, cleanPath); err != nil {
		return "", "", err
	}

	return cleanPath, cleanPath, nil
}

// normalizeAllowedExtensions keeps extension rules consistent for safety checks.
func normalizeAllowedExtensions(allowedExtensions []string) []string {
	normalized := make([]string, 0, len(allowedExtensions))
	for _, ext := range allowedExtensions {
		trimmed := strings.TrimSpace(ext)
		if trimmed == "" {
			continue
		}
		trimmed = strings.ToLower(trimmed)
		trimmed = strings.TrimPrefix(trimmed, ".")
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, "."+trimmed)
	}
	return normalized
}

// normalizeAllowedDirs keeps allowlists comparable for path containment checks.
func normalizeAllowedDirs(allowedDirs []string) []string {
	normalized := make([]string, 0, len(allowedDirs))
	for _, dir := range allowedDirs {
		trimmed := strings.TrimSpace(dir)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, filepath.Clean(trimmed))
	}
	return normalized
}

func formatAllowedExtensions(allowedExtensions []string) string {
	if len(allowedExtensions) == 0 {
		return emptyAllowlistLabel
	}
	normalized := make([]string, 0, len(allowedExtensions))
	for _, ext := range allowedExtensions {
		trimmed := strings.TrimSpace(ext)
		trimmed = strings.TrimPrefix(trimmed, ".")
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return emptyAllowlistLabel
	}
	return strings.Join(normalized, ", ")
}

func formatAllowedDirs(allowedDirs []string) string {
	if len(allowedDirs) == 0 {
		return emptyAllowlistLabel
	}
	normalized := make([]string, 0, len(allowedDirs))
	for _, dir := range allowedDirs {
		trimmed := strings.TrimSpace(dir)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return emptyAllowlistLabel
	}
	return strings.Join(normalized, ", ")
}

func formatHomeAllowedPaths(homeDir string) string {
	if homeDir == "" {
		return "within the home directory (excluding home root and hidden top-level directories)"
	}
	return fmt.Sprintf("within %s (excluding %s and hidden top-level directories)", homeDir, homeDir)
}

func (r *Registrator) allowedPathsSummary() string {
	if len(r.allowedDirs) > 0 {
		return formatAllowedDirs(r.allowedDirs)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return formatHomeAllowedPaths("")
	}
	homeDir = filepath.Clean(homeDir)
	return formatHomeAllowedPaths(homeDir)
}

// prepareSavePath validates the destination path and checks overwrite rules.
func (r *Registrator) prepareSavePath(ctx context.Context, savePath string, override bool) (string, string, error) {
	fullPath, savedPath, err := r.resolveSavePath(ctx, savePath)
	if err != nil {
		return "", "", err
	}
	if override {
		return fullPath, savedPath, nil
	}
	if _, err := os.Stat(fullPath); err == nil {
		return "", "", errors.New("save_path already exists")
	} else if !os.IsNotExist(err) {
		return "", "", r.logError(ctx, fmt.Errorf("check save_path: %w", err))
	}

	return fullPath, savedPath, nil
}

// validateAttachmentExtension blocks unsupported file types to reduce risk.
func (r *Registrator) validateAttachmentExtension(cleanPath string) error {
	allowedExtensions := formatAllowedExtensions(r.allowedExtensions)
	if len(r.allowedExtensions) == 0 {
		return fmt.Errorf("save_path extensions list is empty; allowed extensions: %s", allowedExtensions)
	}
	baseName := strings.ToLower(filepath.Base(cleanPath))
	for _, ext := range r.allowedExtensions {
		if strings.HasSuffix(baseName, ext) {
			return nil
		}
	}
	return fmt.Errorf("save_path extension is not allowed; allowed extensions: %s", allowedExtensions)
}

// validateAttachmentViewExtension blocks unsupported file types for inline viewing.
func (r *Registrator) validateAttachmentViewExtension(fileName string) error {
	allowedExtensions := formatAllowedExtensions(r.allowedViewExts)
	if len(r.allowedViewExts) == 0 {
		return fmt.Errorf("get_content extensions list is empty; allowed extensions: %s", allowedExtensions)
	}
	baseName := strings.ToLower(filepath.Base(fileName))
	for _, ext := range r.allowedViewExts {
		if strings.HasSuffix(baseName, ext) {
			return nil
		}
	}
	return fmt.Errorf("file_name extension is not allowed for get_content; allowed extensions: %s", allowedExtensions)
}

// validateAttachmentDirectory enforces the write scope to prevent unintended writes.
func (r *Registrator) validateAttachmentDirectory(ctx context.Context, cleanPath string) error {
	resolvedPath, err := resolvePathForContainment(cleanPath)
	if err != nil {
		return r.logError(ctx, fmt.Errorf("resolve save_path: %w", err))
	}

	if len(r.allowedDirs) > 0 {
		if ok, err := isWithinAllowedDirs(cleanPath, r.allowedDirs); err != nil {
			return r.logError(ctx, err)
		} else if !ok {
			return fmt.Errorf("save_path must be within allowed directories: %s", formatAllowedDirs(r.allowedDirs))
		}
		if ok, err := isWithinResolvedAllowedDirs(resolvedPath, r.allowedDirs); err != nil {
			return r.logError(ctx, err)
		} else if !ok {
			return fmt.Errorf("save_path must be within allowed directories: %s", formatAllowedDirs(r.allowedDirs))
		}
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return r.logError(ctx, fmt.Errorf("resolve home directory: %w", err))
	}
	homeDir = filepath.Clean(homeDir)
	allowedPaths := formatHomeAllowedPaths(homeDir)
	resolvedHomeDir, err := resolvePathForContainment(homeDir)
	if err != nil {
		return r.logError(ctx, fmt.Errorf("resolve home directory: %w", err))
	}
	if resolvedPath == resolvedHomeDir {
		return fmt.Errorf("save_path must not be home directory; allowed paths: %s", allowedPaths)
	}
	if ok, err := isWithinResolvedRoot(resolvedPath, resolvedHomeDir); err != nil {
		return r.logError(ctx, fmt.Errorf("resolve save_path: %w", err))
	} else if !ok {
		return fmt.Errorf("save_path must be within home directory; allowed paths: %s", allowedPaths)
	}
	relativePath, err := filepath.Rel(resolvedHomeDir, resolvedPath)
	if err != nil {
		return r.logError(ctx, fmt.Errorf("resolve save_path: %w", err))
	}

	segments := strings.Split(relativePath, string(os.PathSeparator))
	if len(segments) > 0 {
		topLevelName := segments[0]
		topLevelPath := filepath.Join(resolvedHomeDir, topLevelName)
		hidden, err := isHiddenTopLevelDir(topLevelName, topLevelPath)
		if err != nil {
			return r.logError(ctx, fmt.Errorf("resolve save_path: %w", err))
		}
		if hidden {
			return fmt.Errorf("save_path must not be within hidden top-level home directory; allowed paths: %s", allowedPaths)
		}
	}

	return nil
}

// logError ensures system errors are recorded with tracker context.
func (r *Registrator) logError(ctx context.Context, err error) error {
	return domain.LogError(ctx, string(domain.ServiceTracker), err)
}

// isWithinAllowedDirs prevents writes outside the explicit allowlist.
func isWithinAllowedDirs(cleanPath string, allowedDirs []string) (bool, error) {
	for _, dir := range allowedDirs {
		if dir == "" {
			continue
		}
		if !filepath.IsAbs(dir) {
			return false, errors.New("allowed directories must be absolute")
		}
		if !strings.EqualFold(filepath.VolumeName(cleanPath), filepath.VolumeName(dir)) {
			continue
		}
		relativePath, err := filepath.Rel(dir, cleanPath)
		if err != nil {
			return false, fmt.Errorf("resolve save_path: %w", err)
		}
		if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
			continue
		}
		return true, nil
	}
	return false, nil
}

// isWithinResolvedAllowedDirs verifies containment after resolving symlinks.
func isWithinResolvedAllowedDirs(resolvedPath string, allowedDirs []string) (bool, error) {
	for _, dir := range allowedDirs {
		if dir == "" {
			continue
		}
		if !filepath.IsAbs(dir) {
			return false, errors.New("allowed directories must be absolute")
		}
		resolvedDir, err := resolvePathForContainment(dir)
		if err != nil {
			return false, fmt.Errorf("resolve save_path: %w", err)
		}
		ok, err := isWithinResolvedRoot(resolvedPath, resolvedDir)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// resolvePathForContainment resolves symlinks for existing parents and rebuilds the full path.
func resolvePathForContainment(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absPath = filepath.Clean(absPath)
	current := absPath
	for {
		_, statErr := os.Lstat(current)
		if statErr == nil {
			return buildResolvedPath(current, absPath)
		}
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", statErr
		}
		current = parent
	}
}

// buildResolvedPath rebuilds the target path with symlinks resolved for an existing parent.
func buildResolvedPath(existingPath, absPath string) (string, error) {
	resolvedCurrent, err := filepath.EvalSymlinks(existingPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(existingPath, absPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedCurrent, rel), nil
}

// isWithinResolvedRoot checks if resolvedPath is inside resolvedRoot.
func isWithinResolvedRoot(resolvedPath, resolvedRoot string) (bool, error) {
	cleanPath := filepath.Clean(resolvedPath)
	cleanRoot := filepath.Clean(resolvedRoot)
	if !strings.EqualFold(filepath.VolumeName(cleanPath), filepath.VolumeName(cleanRoot)) {
		return false, nil
	}
	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return true, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, nil
	}
	return true, nil
}

// writeAttachmentStream writes streamed attachment data to disk.
func (r *Registrator) writeAttachmentStream(fullPath string, override bool, reader io.Reader) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(fullPath), attachmentDirPerm); err != nil {
		return 0, fmt.Errorf("create attachment directory: %w", err)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if override {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	//nolint:gosec // save_path is validated and constrained to allowed locations
	file, err := os.OpenFile(fullPath, flags, attachmentFilePerm)
	if err != nil {
		return 0, fmt.Errorf("open save_path: %w", err)
	}

	bytesWritten, err := io.Copy(file, reader)
	if err != nil {
		_ = file.Close()
		return 0, fmt.Errorf("write attachment: %w", err)
	}
	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("close attachment: %w", err)
	}

	return bytesWritten, nil
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
