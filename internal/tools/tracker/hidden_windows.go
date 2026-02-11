//go:build windows

package tracker

import (
	"errors"
	"os"
	"strings"
	"syscall"
)

// isHiddenTopLevelDir returns true when the first home directory segment is hidden.
func isHiddenTopLevelDir(segmentName, segmentPath string) (bool, error) {
	if strings.HasPrefix(segmentName, ".") {
		return true, nil
	}
	if segmentPath == "" {
		return false, nil
	}
	pathPtr, err := syscall.UTF16PtrFromString(segmentPath)
	if err != nil {
		return false, err
	}
	attrs, err := syscall.GetFileAttributes(pathPtr)
	if err != nil {
		if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) || errors.Is(err, syscall.ERROR_PATH_NOT_FOUND) {
			return false, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
