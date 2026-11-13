package oauthfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// TestAcquire covers token normalization, re-reading a replaced file, and safe failures.
func TestAcquire(t *testing.T) {
	t.Parallel()

	const token = "y0__xTestToken-123_abc"
	path := filepath.Join(t.TempDir(), "token")
	source := New(path)

	require.NoError(t, os.WriteFile(path, []byte("\xef\xbb\xbf"+token+"\r\n"), 0o600))
	got, err := source.Acquire(t.Context(), "ignored-profile")
	require.NoError(t, err)
	require.Equal(t, token, got)

	require.NoError(t, os.WriteFile(path, []byte("replaced"), 0o600))
	got, err = source.Acquire(t.Context(), "")
	require.NoError(t, err)
	require.Equal(t, "replaced", got)

	for name, content := range map[string]string{
		"empty":     " \r\n",
		"two lines": "first\nsecond",
		"inner gap": "first second",
		"non-ascii": "токен",
	} {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600), name)
		got, err = source.Acquire(t.Context(), "")
		require.ErrorAs(t, err, &domain.AuthenticationError{}, name)
		require.Empty(t, got, name)
		if content != " \r\n" {
			require.NotContains(t, err.Error(), content, name)
		}
	}

	got, err = New(filepath.Join(t.TempDir(), "missing")).Acquire(t.Context(), "")
	require.ErrorAs(t, err, &domain.AuthenticationError{})
	require.ErrorIs(t, err, os.ErrNotExist)
	require.Empty(t, got)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	got, err = source.Acquire(ctx, "")
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, got)
}
