//go:build windows

package tracker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsWithinAllowedDirs_VolumeMismatch ensures volume mismatches are treated as non-matches to avoid false errors.
func TestIsWithinAllowedDirs_VolumeMismatch(t *testing.T) {
	t.Parallel()
	allowedDirs := []string{`C:\allowed`}
	cleanPath := `D:\target\file.txt`

	ok, err := isWithinAllowedDirs(cleanPath, allowedDirs)
	require.NoError(t, err)
	assert.False(t, ok)
}
