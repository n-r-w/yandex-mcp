package attachments

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Write streams data to a path validated by PrepareSavePath and returns the byte count on success.
func (s *Service) Write(fullPath string, override bool, reader io.Reader) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(fullPath), attachmentDirPerm); err != nil {
		return 0, fmt.Errorf("create attachment directory: %w", err)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if override {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	//nolint:gosec // save_path is validated and constrained to allowed locations
	file, err := os.OpenFile(fullPath, flags, attachmentFilePerm)
	if err != nil {
		return 0, fmt.Errorf("open save_path: %w", err)
	}

	bytesWritten, err := io.Copy(file, reader)
	if err != nil {
		_ = file.Close()
		return 0, fmt.Errorf("write attachment: %w", err)
	}
	if closeErr := file.Close(); closeErr != nil {
		return 0, fmt.Errorf("close attachment: %w", closeErr)
	}

	return bytesWritten, nil
}
