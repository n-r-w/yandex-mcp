//go:build !windows

package tracker

import "strings"

// isHiddenTopLevelDir returns true when the first home directory segment is hidden.
func isHiddenTopLevelDir(segmentName, _ string) (bool, error) {
	return strings.HasPrefix(segmentName, "."), nil
}
