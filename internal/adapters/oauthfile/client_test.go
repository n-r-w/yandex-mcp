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

// TestAcquireProfile selects different token files by profile name and rejects profiles
// that have no file, leave the directory, or arrive when no directory is configured.
func TestAcquireProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "work"), []byte("work-token\r\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "personal.2"), []byte("personal-token"), 0o600))
	require.NoError(t, os.WriteFile(outside, []byte("outside-token"), 0o600))
	source := NewProfiles(dir)

	for profile, want := range map[string]string{"work": "work-token", "personal.2": "personal-token"} {
		got, err := source.Acquire(t.Context(), profile)
		require.NoError(t, err, profile)
		require.Equal(t, want, got, profile)
	}

	got, err := source.Acquire(t.Context(), "absent")
	require.ErrorAs(t, err, &domain.AuthenticationError{})
	require.ErrorIs(t, err, errUnknownProfile)
	require.ErrorContains(t, err, `"absent"`)
	require.Empty(t, got)

	relative, err := filepath.Rel(dir, outside)
	require.NoError(t, err)
	escaping := []string{"", ".", "..", ".hidden", relative, outside, "work/../work", `work\..\work`, "work:stream"}
	for _, profile := range escaping {
		got, err = source.Acquire(t.Context(), profile)
		require.ErrorAs(t, err, &domain.AuthenticationError{}, profile)
		require.ErrorIs(t, err, errInvalidProfile, profile)
		require.Empty(t, got, profile)
		require.NotContains(t, err.Error(), "outside-token", profile)
	}

	got, err = NewProfiles("").Acquire(t.Context(), "work")
	require.ErrorAs(t, err, &domain.AuthenticationError{})
	require.ErrorIs(t, err, errNoProfilesDir)
	require.Empty(t, got)
}

// TestAcquireFileIgnoresProfile keeps the single-file source independent of the profile name.
func TestAcquireFileIgnoresProfile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte("file-token"), 0o600))

	got, err := New(path).Acquire(t.Context(), "../any/profile")
	require.NoError(t, err)
	require.Equal(t, "file-token", got)
}
