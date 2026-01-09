// Package helpers provides shared utilities for MCP tools.
package helpers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// ErrorLogWrapper logs the error with context and returns it.
func ErrorLogWrapper(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	slog.ErrorContext(ctx, "tools error", "error", err)
	return err
}

// ToSafeError converts errors to safe tool errors that do not leak sensitive information.
func ToSafeError(ctx context.Context, err error, serviceName string) (errOut error) {
	defer func() {
		errOut = ErrorLogWrapper(ctx, errOut)
	}()

	var upstreamErr domain.UpstreamError
	if errors.As(err, &upstreamErr) {
		return fmt.Errorf("%s %s: %s (HTTP %d)",
			upstreamErr.Service,
			upstreamErr.Operation,
			upstreamErr.Message,
			upstreamErr.HTTPStatus,
		)
	}

	// For other errors, extract useful context from error message
	// while avoiding sensitive data leakage
	errMsg := err.Error()

	// Check for common error patterns that are safe to expose
	safePrefixes := []string{
		"decode response:",
		"read response body:",
		"parse base url:",
		"create request:",
		"marshal request body:",
		"execute request:",
		"get token:",
	}

	safeContains := []string{
		"unsupported protocol scheme",
		"unprocessable entity",
	}

	for _, prefix := range safePrefixes {
		if strings.HasPrefix(strings.ToLower(errMsg), prefix) {
			// These are technical error messages that don't contain sensitive data
			return fmt.Errorf("%s: %s", serviceName, errMsg)
		}
	}

	for _, substr := range safeContains {
		if strings.Contains(strings.ToLower(errMsg), substr) {
			// These are known safe error messages
			return fmt.Errorf("%s: %s", serviceName, errMsg)
		}
	}

	slog.ErrorContext(ctx, "unhandled error pattern", "error", errMsg)

	// For any other errors, return a generic safe message to avoid leaking sensitive data
	return fmt.Errorf("%s: internal error", serviceName)
}

// ConvertFilterToStringMap converts a map[string]any filter to map[string]string.
func ConvertFilterToStringMap(ctx context.Context, filter map[string]any) (map[string]string, error) {
	if filter == nil {
		return nil, nil //nolint:nilnil // nil filter means no filter, not an error
	}
	result := make(map[string]string, len(filter))
	for k, v := range filter {
		s, ok := v.(string)
		if !ok {
			return nil, ErrorLogWrapper(ctx, fmt.Errorf("filter value for key %q must be a string, got %T", k, v))
		}
		result[k] = s
	}
	return result, nil
}
