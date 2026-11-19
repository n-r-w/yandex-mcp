package attachments

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// resolveSavePath validates the absolute save path against attachment safety rules.
func (s *Service) resolveSavePath(ctx context.Context, service domain.Service, savePath string) (string, error) {
	cleanPath := filepath.Clean(savePath)
	if !filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("save_path must be absolute; allowed paths: %s", s.allowedPathsSummary())
	}
	if cleanPath == "." || cleanPath == string(os.PathSeparator) {
		return "", fmt.Errorf("save_path must point to a file; allowed paths: %s", s.allowedPathsSummary())
	}
	if validateErr := s.validateAttachmentExtension(cleanPath); validateErr != nil {
		return "", validateErr
	}
	if validateErr := s.validateAttachmentDirectory(ctx, service, cleanPath); validateErr != nil {
		return "", validateErr
	}

	return cleanPath, nil
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

// formatAllowedExtensions lists permitted suffixes for validation errors.
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

// formatAllowedDirs lists permitted directories for validation errors.
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

// formatDefaultAllowedPaths describes the write scope without a directory allowlist.
func formatDefaultAllowedPaths(homeDir string) string {
	tempDir := filepath.Clean(os.TempDir())
	if homeDir == "" {
		return "within the home directory (excluding home root and hidden top-level directories), or within " + tempDir
	}
	return fmt.Sprintf(
		"within %s (excluding %s and hidden top-level directories), or within %s",
		homeDir, homeDir, tempDir,
	)
}

// allowedPathsSummary describes the configured write scope for validation errors.
func (s *Service) allowedPathsSummary() string {
	if len(s.allowedDirs) > 0 {
		return formatAllowedDirs(s.allowedDirs)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return formatDefaultAllowedPaths("")
	}
	homeDir = filepath.Clean(homeDir)
	return formatDefaultAllowedPaths(homeDir)
}

// PrepareSavePath validates the destination and returns the cleaned path to open and report.
func (s *Service) PrepareSavePath(
	ctx context.Context,
	service domain.Service,
	savePath string,
	override bool,
) (string, error) {
	fullPath, err := s.resolveSavePath(ctx, service, savePath)
	if err != nil {
		return "", err
	}
	if override {
		return fullPath, nil
	}
	if _, statErr := os.Stat(fullPath); statErr == nil {
		return "", errors.New("save_path already exists")
	} else if !os.IsNotExist(statErr) {
		return "", domain.LogError(ctx, string(service), fmt.Errorf("check save_path: %w", statErr))
	}

	return fullPath, nil
}

// validateAttachmentExtension blocks unsupported file types to reduce risk.
func (s *Service) validateAttachmentExtension(cleanPath string) error {
	allowedExtensions := formatAllowedExtensions(s.allowedExtensions)
	if len(s.allowedExtensions) == 0 {
		return fmt.Errorf("save_path extensions list is empty; allowed extensions: %s", allowedExtensions)
	}
	baseName := strings.ToLower(filepath.Base(cleanPath))
	for _, ext := range s.allowedExtensions {
		if strings.HasSuffix(baseName, ext) {
			return nil
		}
	}
	return fmt.Errorf(
		"save_path extension %q is not allowed; allowed extensions: %s",
		rejectedPathExtension(cleanPath),
		allowedExtensions,
	)
}

// rejectedPathExtension returns the suffix that failed extension allowlist validation.
func rejectedPathExtension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return emptyAllowlistLabel
	}
	return strings.ToLower(ext)
}

// ValidateViewExtension checks the inline suffix and names the caller's filename source on rejection.
func (s *Service) ValidateViewExtension(fileName, fileNameSource string) error {
	allowedExtensions := formatAllowedExtensions(s.allowedViewExts)
	if len(s.allowedViewExts) == 0 {
		return fmt.Errorf("get_content extensions list is empty; allowed extensions: %s", allowedExtensions)
	}
	baseName := strings.ToLower(filepath.Base(fileName))
	for _, ext := range s.allowedViewExts {
		if strings.HasSuffix(baseName, ext) {
			return nil
		}
	}
	return fmt.Errorf(
		"%s extension is not allowed for get_content; allowed extensions: %s",
		fileNameSource,
		allowedExtensions,
	)
}

// validateAttachmentDirectory enforces the write scope to prevent unintended writes.
func (s *Service) validateAttachmentDirectory(ctx context.Context, service domain.Service, cleanPath string) error {
	resolvedPath, err := resolvePathForContainment(cleanPath)
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve save_path: %w", err))
	}

	if len(s.allowedDirs) > 0 {
		return s.validateWithinAllowedDirs(ctx, service, cleanPath, resolvedPath)
	}

	withinTempDir, err := isWithinTempDir(resolvedPath)
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve temporary directory: %w", err))
	}
	if withinTempDir {
		return nil
	}

	return s.validateWithinHomeDir(ctx, service, resolvedPath)
}

// validateWithinAllowedDirs enforces both direct and resolved path containment for explicit allowlists.
func (s *Service) validateWithinAllowedDirs(
	ctx context.Context,
	service domain.Service,
	cleanPath, resolvedPath string,
) error {
	allowedDirs := formatAllowedDirs(s.allowedDirs)
	withinAllowedDirs, err := isWithinAllowedDirs(cleanPath, s.allowedDirs)
	if err != nil {
		return domain.LogError(ctx, string(service), err)
	}
	if !withinAllowedDirs {
		return fmt.Errorf("save_path must be within allowed directories: %s", allowedDirs)
	}

	withinResolvedAllowedDirs, err := isWithinResolvedAllowedDirs(resolvedPath, s.allowedDirs)
	if err != nil {
		return domain.LogError(ctx, string(service), err)
	}
	if !withinResolvedAllowedDirs {
		return fmt.Errorf("save_path must be within allowed directories: %s", allowedDirs)
	}

	return nil
}

// validateWithinHomeDir enforces containment rules when the home directory fallback policy is active.
func (s *Service) validateWithinHomeDir(ctx context.Context, service domain.Service, resolvedPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve home directory: %w", err))
	}
	homeDir = filepath.Clean(homeDir)
	allowedPaths := formatDefaultAllowedPaths(homeDir)
	resolvedHomeDir, err := resolvePathForContainment(homeDir)
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve home directory: %w", err))
	}
	if resolvedPath == resolvedHomeDir {
		return fmt.Errorf("save_path must not be home directory; allowed paths: %s", allowedPaths)
	}
	withinResolvedRoot, err := isWithinResolvedRoot(resolvedPath, resolvedHomeDir)
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve save_path: %w", err))
	}
	if !withinResolvedRoot {
		return fmt.Errorf("save_path must be within home directory; allowed paths: %s", allowedPaths)
	}

	return s.validateHomeTopLevelDir(ctx, service, resolvedHomeDir, resolvedPath, allowedPaths)
}

// validateHomeTopLevelDir rejects writes into hidden top-level home subdirectories.
func (s *Service) validateHomeTopLevelDir(
	ctx context.Context,
	service domain.Service,
	resolvedHomeDir string,
	resolvedPath string,
	allowedPaths string,
) error {
	relativePath, err := filepath.Rel(resolvedHomeDir, resolvedPath)
	if err != nil {
		return domain.LogError(ctx, string(service), fmt.Errorf("resolve save_path: %w", err))
	}

	segments := strings.Split(relativePath, string(os.PathSeparator))
	if len(segments) > 0 {
		topLevelName := segments[0]
		topLevelPath := filepath.Join(resolvedHomeDir, topLevelName)
		hidden, hiddenErr := isHiddenTopLevelDir(topLevelName, topLevelPath)
		if hiddenErr != nil {
			return domain.LogError(ctx, string(service), fmt.Errorf("resolve save_path: %w", hiddenErr))
		}
		if hidden {
			return fmt.Errorf(
				"save_path must not be within hidden top-level home directory; allowed paths: %s",
				allowedPaths,
			)
		}
	}

	return nil
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
		if relativePath == "." ||
			relativePath == parentDirMarker ||
			strings.HasPrefix(relativePath, parentDirMarker+string(os.PathSeparator)) {
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

// isWithinTempDir reports whether resolved save path is inside system temporary directory.
func isWithinTempDir(resolvedPath string) (bool, error) {
	resolvedTempDir, err := resolvePathForContainment(os.TempDir())
	if err != nil {
		return false, err
	}
	if resolvedPath == resolvedTempDir {
		return false, nil
	}
	return isWithinResolvedRoot(resolvedPath, resolvedTempDir)
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
	if rel == parentDirMarker || strings.HasPrefix(rel, parentDirMarker+string(os.PathSeparator)) {
		return false, nil
	}
	return true, nil
}
