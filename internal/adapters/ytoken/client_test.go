package ytoken

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Cache tests use a source mock. Inputs cover cache hits, expiry, forced refresh,
// and refresh failure. A failed refresh must return its error, not the old token.
func TestProviderCache(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		source := NewMockITokenSource(gomock.NewController(t))
		provider := New(source, "work", time.Hour)
		defer provider.Close()
		failure := errors.New("authentication failed")
		gomock.InOrder(
			source.EXPECT().Acquire(gomock.Any(), "work").Return("first", nil),
			source.EXPECT().Acquire(gomock.Any(), "work").Return("expired", nil),
			source.EXPECT().Acquire(gomock.Any(), "work").Return("forced", nil),
			source.EXPECT().Acquire(gomock.Any(), "work").Return("", failure),
		)
		for range 2 {
			token, err := provider.Token(t.Context(), false)
			require.NoError(t, err)
			require.Equal(t, "first", token)
		}
		time.Sleep(time.Hour)
		token, err := provider.Token(t.Context(), false)
		require.NoError(t, err)
		require.Equal(t, "expired", token)
		token, err = provider.Token(t.Context(), true)
		require.NoError(t, err)
		require.Equal(t, "forced", token)
		token, err = provider.Token(t.Context(), true)
		require.ErrorIs(t, err, failure)
		require.Empty(t, token)
	})
}

// Concurrent callers share the source but cancel independently, including the
// first caller. Completion populates the cache. The only dependency is a mock.
func TestProviderWaiters(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		source := NewMockITokenSource(gomock.NewController(t))
		provider := New(source, "work", time.Hour)
		defer provider.Close()
		release := make(chan struct{})
		source.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(func(ctx context.Context, _ string) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-release:
				return "token", nil
			}
		})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		first := make(chan error, 1)
		go func() { _, err := provider.Token(ctx, false); first <- err }()
		synctest.Wait()
		second := make(chan string, 1)
		go func() { token, err := provider.Token(t.Context(), false); assert.NoError(t, err); second <- token }()
		synctest.Wait()
		cancel()
		require.ErrorIs(t, <-first, context.Canceled)
		close(release)
		require.Equal(t, "token", <-second)
		token, err := provider.Token(t.Context(), false)
		require.NoError(t, err)
		require.Equal(t, "token", token)
	})
}

// Canceling the final provider waiter must cancel its source request.
func TestProviderLastWaiter(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		source := NewMockITokenSource(gomock.NewController(t))
		provider := New(source, "work", time.Hour)
		defer provider.Close()
		stopped := false
		source.EXPECT().Acquire(gomock.Any(), "work").DoAndReturn(func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done()
			stopped = true
			return "", ctx.Err()
		})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		go func() { _, _ = provider.Token(ctx, false) }()
		synctest.Wait()
		cancel()
		synctest.Wait()
		require.True(t, stopped)
	})
}
