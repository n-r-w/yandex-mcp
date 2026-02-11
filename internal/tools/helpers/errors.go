// Package helpers provides shared utilities for MCP tools.
package helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// ToSafeError converts errors to safe tool errors that do not leak sensitive information.
func ToSafeError(ctx context.Context, serviceName domain.Service, err error) (errOut error) {
	if err == nil {
		return nil
	}

	var upstreamErr domain.UpstreamError
	if errors.As(err, &upstreamErr) {
		return fmt.Errorf("%s %s: %s (HTTP %d)",
			upstreamErr.Service,
			upstreamErr.Operation,
			upstreamErr.Message,
			upstreamErr.HTTPStatus,
		)
	}

	errMsg := err.Error()
	lowerMsg := strings.ToLower(errMsg)

	if isSafeError(lowerMsg) {
		return fmt.Errorf("%s: %s", serviceName, errMsg)
	}

	_ = domain.LogError(ctx, string(serviceName), err)

	return fmt.Errorf("%s: internal error", serviceName)
}

// isSafeError checks if the error message matches known safe patterns.
func isSafeError(lowerMsg string) bool {
	for _, prefix := range safePrefixes {
		if strings.HasPrefix(lowerMsg, prefix) {
			return true
		}
	}

	for _, substr := range safeContains {
		if strings.Contains(lowerMsg, substr) {
			return true
		}
	}

	return false
}
