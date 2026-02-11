// Package token provides IAM token acquisition using the yc CLI.
package token

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/n-r-w/singleflight/v2"
	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Provider implements ITokenProvider with caching and single-flight behavior.
type Provider struct {
	executor      ICommandExecutor
	refreshPeriod time.Duration
	nowFunc       func() time.Time

	mu          sync.RWMutex
	cachedToken string
	refreshedAt time.Time

	// single-flight group for token refresh operations
	sf singleflight.Group[string, string]

	tokenRegex *regexp.Regexp
}

// Compile-time interface assertions.
var _ apihelpers.ITokenProvider = (*Provider)(nil)

// NewProvider creates a new token provider.
func NewProvider(cfg *config.Config) *Provider {
	//nolint:exhaustruct // cache and sync fields intentionally start with zero values
	return &Provider{
		executor:      newCommandExecutor(),
		refreshPeriod: cfg.IAMTokenRefreshPeriod,
		nowFunc:       time.Now,
		tokenRegex:    regexp.MustCompile(tokenRegexPattern),
	}
}

// setNowFunc sets the time function for testing. Not thread-safe; call before use.
func (p *Provider) setNowFunc(fn func() time.Time) {
	p.nowFunc = fn
}

// setExecutor sets the command executor for testing. Not thread-safe; call before use.
func (p *Provider) setExecutor(exec ICommandExecutor) {
	p.executor = exec
}

// Token returns a cached IAM token or fetches a new one if cache is stale.
func (p *Provider) Token(ctx context.Context, forceRefresh bool) (string, error) {
	if !forceRefresh {
		// Fast path: check if cached token is still valid
		if token, ok := p.getCachedToken(); ok {
			return token, nil
		}
	}

	return p.refreshToken(ctx)
}

// getCachedToken returns the cached token if it exists and is not expired.
func (p *Provider) getCachedToken() (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cachedToken == "" {
		return "", false
	}

	if p.nowFunc().Sub(p.refreshedAt) >= p.refreshPeriod {
		return "", false
	}

	return p.cachedToken, true
}

// refreshToken fetches a new token from yc CLI with single-flight coordination.
func (p *Provider) refreshToken(ctx context.Context) (string, error) {
	// Use singleflight to ensure only one refresh happens at a time
	result, _, err := p.sf.Do(ctx, "token", p.doRefresh)
	if err != nil {
		return "", err
	}

	return result, nil
}

// doRefresh performs the actual token refresh.
func (p *Provider) doRefresh(ctx context.Context) (string, error) {
	token, err := p.executeYC(ctx)
	if err != nil {
		return "", err
	}

	// Update cache
	p.mu.Lock()
	p.cachedToken = token
	p.refreshedAt = p.nowFunc()
	p.mu.Unlock()

	return token, nil
}

// executeYC runs the yc CLI command and extracts the IAM token using regex.
func (p *Provider) executeYC(ctx context.Context) (string, error) {
	output, err := p.executor.Execute(ctx, "yc", "iam", "create-token")
	if err != nil {
		// Check context errors before sanitizing (sanitizeError breaks error chain)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("token fetch canceled or timed out: %w", err)
		}

		return "", p.logError(ctx, fmt.Errorf("%w: %s", errTokenFetchFailed, p.sanitizeError(err).Error()))
	}

	if len(output) == 0 {
		return "", p.logError(ctx, errEmptyToken)
	}

	match := p.tokenRegex.Find(output)
	if match == nil {
		return "", p.logError(ctx, errTokenNotFound)
	}

	return string(match), nil
}
