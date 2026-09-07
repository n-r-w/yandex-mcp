package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// Token diagnostics must survive API wrapping and appear in both MCP output and
// structured stderr logs. The log buffer replaces stderr; no external systems run.
//
//nolint:paralleltest // the process-wide slog logger must be restored before other tests run
func TestAuthenticationErrorPreservesDiagnostics(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	original := "exit status 1: original yc diagnostic"
	err := fmt.Errorf("get token: %w", domain.AuthenticationError{Err: errors.New(original)})
	result := ToSafeError(t.Context(), domain.ServiceWiki, err)
	require.ErrorContains(t, result, original)
	require.Contains(t, logs.String(), original)
}
