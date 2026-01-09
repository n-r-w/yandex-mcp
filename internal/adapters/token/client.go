// Package token provides IAM token acquisition using the yc CLI.
package token

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

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

	// in-flight coordination for single-flight refresh
	inflight    bool
	inflightCh  chan struct{}
	inflightErr error

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
func (p *Provider) Token(ctx context.Context) (string, error) {
	// Fast path: check if cached token is still valid
	if token, ok := p.getCachedToken(); ok {
		return token, nil
	}

	return p.refreshToken(ctx, false)
}

// ForceRefresh discards any cached token and fetches a new one.
func (p *Provider) ForceRefresh(ctx context.Context) (string, error) {
	return p.refreshToken(ctx, true)
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
func (p *Provider) refreshToken(ctx context.Context, force bool) (string, error) {
	p.mu.Lock()

	// Check if another goroutine is already fetching
	if p.inflight {
		ch := p.inflightCh
		p.mu.Unlock()
		// Wait for in-flight request to complete
		select {
		case <-ch:
			return p.getInflightResult()
		case <-ctx.Done():
			return "", p.logError(ctx, fmt.Errorf("token refresh: %w", ctx.Err()))
		}
	}

	// Double-check: maybe token was refreshed while we waited for lock
	// Skip this check for force refresh
	if !force && p.cachedToken != "" && p.nowFunc().Sub(p.refreshedAt) < p.refreshPeriod {
		token := p.cachedToken
		p.mu.Unlock()
		return token, nil
	}

	// Start in-flight request
	p.inflight = true
	p.inflightCh = make(chan struct{})
	p.inflightErr = nil
	p.mu.Unlock()

	// Execute yc command
	token, err := p.executeYC(ctx)

	// Update cache and signal completion
	p.mu.Lock()
	if err == nil {
		p.cachedToken = token
		p.refreshedAt = p.nowFunc()
	} else {
		p.inflightErr = err
	}
	p.inflight = false
	close(p.inflightCh)
	p.mu.Unlock()

	if err != nil {
		return "", p.logError(ctx, err)
	}

	return token, nil
}

// getInflightResult returns the result of the in-flight refresh operation.
// Returns cached token if refresh succeeded, or the refresh error if it failed.
func (p *Provider) getInflightResult() (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.inflightErr != nil {
		return "", p.inflightErr
	}

	if p.cachedToken == "" {
		return "", errEmptyToken
	}

	return p.cachedToken, nil
}

// executeYC runs the yc CLI command and extracts the IAM token using regex.
func (p *Provider) executeYC(ctx context.Context) (string, error) {
	output, err := p.executor.Execute(ctx, "yc", "iam", "create-token")
	if err != nil {
		// Check context errors before sanitizing (sanitizeError breaks error chain)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", p.logError(ctx, fmt.Errorf("token fetch canceled or timed out: %w", err))
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
