package attachments

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// TestWriteRejectsFileCreatedAfterPreflight verifies that opening the destination enforces overwrite permission.
func TestWriteRejectsFileCreatedAfterPreflight(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	delivery := New([]string{"txt"}, nil, []string{dir})
	path, err := delivery.PrepareSavePath(t.Context(), domain.ServiceTracker, filepath.Join(dir, "report.txt"), false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("original"), 0o600))

	_, err = delivery.Write(path, false, strings.NewReader("replacement"))
	require.ErrorIs(t, err, fs.ErrExist)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "original", string(data))
}

// TestWriteOverwritesSymlinkTarget verifies that writing through a permitted link preserves the link.
func TestWriteOverwritesSymlinkTarget(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "archive.txt")
	link := filepath.Join(dir, "report.txt")
	require.NoError(t, os.WriteFile(target, []byte("original contents"), 0o600))
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	delivery := New([]string{"txt"}, nil, []string{dir})
	path, err := delivery.PrepareSavePath(t.Context(), domain.ServiceTracker, link, true)
	require.NoError(t, err)

	written, err := delivery.Write(path, true, strings.NewReader("new"))
	require.NoError(t, err)
	require.EqualValues(t, 3, written)
	linkTarget, err := os.Readlink(link)
	require.NoError(t, err)
	require.Equal(t, target, linkTarget)
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
}

// TestWriteReportsCopyFailure verifies that the caller receives the stream error after a partial write.
func TestWriteReportsCopyFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	delivery := New([]string{"txt"}, nil, []string{dir})
	path, err := delivery.PrepareSavePath(t.Context(), domain.ServiceTracker, filepath.Join(dir, "report.txt"), false)
	require.NoError(t, err)
	readErr := errors.New("read failed")
	reader := io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(readErr))

	_, err = delivery.Write(path, false, reader)
	require.ErrorIs(t, err, readErr)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "partial", string(data))
}
