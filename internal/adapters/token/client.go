// Package token provides IAM token acquisition using the yc CLI.
package token

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/adapters/tracker"
	"github.com/n-r-w/yandex-mcp/internal/adapters/wiki"
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
}

// Compile-time interface assertions.
var (
	_ tracker.ITokenProvider = (*Provider)(nil)
	_ wiki.ITokenProvider    = (*Provider)(nil)
)

// NewProvider creates a new token provider.
func NewProvider(cfg *config.Config) *Provider {
	return &Provider{
		executor:      newCommandExecutor(),
		refreshPeriod: cfg.IAMTokenRefreshPeriod,
		nowFunc:       time.Now,
		mu:            sync.RWMutex{},
		cachedToken:   "",
		refreshedAt:   time.Time{},
		inflight:      false,
		inflightCh:    nil,
		inflightErr:   nil,
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
			return "", errorLogWrapper(ctx, fmt.Errorf("token refresh: %w", ctx.Err()))
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
		return "", errorLogWrapper(ctx, err)
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
			return "", errorLogWrapper(ctx, fmt.Errorf("token fetch canceled or timed out: %w", err))
		}

		//nolint:errorlint // wrap for sentinel comparison
		return "", errorLogWrapper(ctx, fmt.Errorf("%w: %v", errTokenFetchFailed, sanitizeError(err)))
	}

	if len(output) == 0 {
		return "", errorLogWrapper(ctx, errEmptyToken)
	}

	// Extract token using regex
	re := regexp.MustCompile(tokenRegexPattern)
	match := re.Find(output)
	if match == nil {
		return "", errorLogWrapper(ctx, errTokenNotFound)
	}

	return string(match), nil
}
