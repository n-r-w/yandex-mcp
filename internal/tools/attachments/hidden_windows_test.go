//go:build windows

package attachments

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsHiddenTopLevelDir_WindowsHiddenAttribute checks a directory hidden by its Windows file attribute.
func TestIsHiddenTopLevelDir_WindowsHiddenAttribute(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	hiddenDir := filepath.Join(baseDir, "HiddenTop")
	require.NoError(t, os.MkdirAll(hiddenDir, 0o755))

	pathPtr, err := syscall.UTF16PtrFromString(hiddenDir)
	require.NoError(t, err)
	require.NoError(t, syscall.SetFileAttributes(pathPtr, syscall.FILE_ATTRIBUTE_HIDDEN))

	hidden, err := isHiddenTopLevelDir("HiddenTop", hiddenDir)
	require.NoError(t, err)
	assert.True(t, hidden)
}
