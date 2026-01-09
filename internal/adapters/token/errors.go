package token

import (
	"context"
	"errors"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

var (
	errTokenFetchFailed = errors.New("failed to fetch IAM token")
	errEmptyToken       = errors.New("empty token received from yc")
	errTokenNotFound    = errors.New("token not found in yc output")
)

// sanitizeError redacts IAM token patterns from error messages.
func (p *Provider) sanitizeError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := p.tokenRegex.ReplaceAllString(err.Error(), "[REDACTED_TOKEN]")
	return errors.New(errMsg)
}

func (p *Provider) logError(ctx context.Context, err error) error {
	return domain.LogError(ctx, "token", err)
}
