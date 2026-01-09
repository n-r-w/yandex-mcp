package token

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
)

var (
	errTokenFetchFailed = errors.New("failed to fetch IAM token")
	errEmptyToken       = errors.New("empty token received from yc")
	errTokenNotFound    = errors.New("token not found in yc output")
)

func errorLogWrapper(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	slog.ErrorContext(ctx, "token adapter error", "error", err)
	return err
}

// sanitizeError removes sensitive data (tokens, command output) from error messages.
func sanitizeError(err error) error {
	if err == nil {
		return nil
	}

	// Get the error message
	errMsg := err.Error()

	// Remove any IAM tokens (pattern: t1.xxx.yyy)
	re := regexp.MustCompile(tokenRegexPattern)
	errMsg = re.ReplaceAllString(errMsg, "[REDACTED_TOKEN]")

	// Return a new error with sanitized message
	// This prevents leaking sensitive command output
	return errors.New(errMsg)
}
