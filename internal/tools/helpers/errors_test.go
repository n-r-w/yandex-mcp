package helpers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

func TestWrapError(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("connection failed: Authorization header: Bearer secret-token-123")

	result := WrapError(t.Context(), domain.ServiceTracker, originalErr)

	require.EqualError(t, result, "tracker: "+originalErr.Error())
	require.ErrorIs(t, result, originalErr)
}

func TestWrapError_WithNilError(t *testing.T) {
	t.Parallel()

	result := WrapError(t.Context(), "test-service", nil)

	require.NoError(t, result)
}
