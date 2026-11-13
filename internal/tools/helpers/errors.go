// Package helpers provides shared utilities for MCP tools.
package helpers

import (
	"context"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// WrapError adds the service name while preserving the original error.
func WrapError(ctx context.Context, serviceName domain.Service, err error) error {
	if err == nil {
		return nil
	}

	_ = domain.LogError(ctx, string(serviceName), err)

	return fmt.Errorf("%s: %w", serviceName, err)
}
