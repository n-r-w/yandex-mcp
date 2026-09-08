// Package ytoken caches IAM tokens obtained from an injected source.
package ytoken

import (
	"context"
	"sync"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/adapters/authwait"
)

// Provider caches tokens and shares active acquisitions within one MCP process.
type Provider struct {
	source        ITokenSource
	profile       string
	refreshPeriod time.Duration
	nowFunc       func() time.Time
	mu            sync.RWMutex
	cachedToken   string
	refreshedAt   time.Time
	acquisitions  authwait.Group
}

var _ apihelpers.ITokenProvider = (*Provider)(nil)

// New constructs a provider. The caller must Close it during shutdown.
func New(source ITokenSource, profile string, refreshPeriod time.Duration) *Provider {
	//nolint:exhaustruct_v5 // synchronization and cache start with zero values
	return &Provider{source: source, profile: profile, refreshPeriod: refreshPeriod, nowFunc: time.Now}
}

// Token returns a cached token or waits for a fresh token.
func (p *Provider) Token(ctx context.Context, forceRefresh bool) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !forceRefresh {
		if token, ok := p.getCachedToken(); ok {
			return token, nil
		}
	}
	return p.acquisitions.Do(ctx, p.profile, func(shared context.Context) (string, error) {
		// A prior acquisition can finish between the fast path and joining the group.
		if !forceRefresh {
			if token, ok := p.getCachedToken(); ok {
				return token, nil
			}
		}
		token, err := p.source.Acquire(shared, p.profile)
		if err != nil {
			return "", err
		}
		p.mu.Lock()
		p.cachedToken, p.refreshedAt = token, p.nowFunc()
		p.mu.Unlock()
		return token, nil
	})
}

func (p *Provider) getCachedToken() (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cachedToken, p.cachedToken != "" && p.nowFunc().Sub(p.refreshedAt) < p.refreshPeriod
}

// Close cancels owned acquisitions and waits for source cleanup.
func (p *Provider) Close() { p.acquisitions.Close() }
