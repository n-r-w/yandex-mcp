package authwait

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two callers share a result; canceling the first must not cancel the acquisition.
// Channels control completion. No external dependencies are used.
func TestIndependentCancellation(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var group Group
		defer group.Close()
		first, cancel := context.WithCancel(t.Context())
		defer cancel()
		release := make(chan struct{})
		calls := 0
		acquire := func(ctx context.Context) (string, error) {
			calls++
			select {
			case <-release:
				return "token", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		firstResult := make(chan error, 1)
		go func() { _, err := group.Do(first, "profile", acquire); firstResult <- err }()
		synctest.Wait()
		secondResult := make(chan string, 1)
		go func() {
			token, err := group.Do(t.Context(), "profile", acquire)
			assert.NoError(t, err)
			secondResult <- token
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		require.ErrorIs(t, <-firstResult, context.Canceled)
		require.Equal(t, 1, calls)
		close(release)
		require.Equal(t, "token", <-secondResult)
	})
}

// The last cancellation stops acquisition, but a new caller must wait for cleanup.
func TestLastWaiterAndProcessExit(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var group Group
		defer group.Close()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		exited := make(chan struct{})
		canceled := false
		go func() {
			_, _ = group.Do(ctx, "profile", func(ctx context.Context) (string, error) {
				<-ctx.Done()
				canceled = true
				<-exited
				return "", ctx.Err()
			})
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		require.True(t, canceled)
		started := false
		result := make(chan string, 1)
		go func() {
			token, err := group.Do(
				t.Context(),
				"profile",
				func(context.Context) (string, error) { started = true; return "next", nil },
			)
			assert.NoError(t, err)
			result <- token
		}()
		synctest.Wait()
		require.False(t, started)
		close(exited)
		require.Equal(t, "next", <-result)
	})
}

// Closing an owner cancels active acquisitions and rejects later calls.
func TestClose(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var group Group
		stopped := false
		go func() {
			_, _ = group.Do(
				t.Context(),
				"profile",
				func(ctx context.Context) (string, error) { <-ctx.Done(); stopped = true; return "", ctx.Err() },
			)
		}()
		synctest.Wait()
		group.Close()
		require.True(t, stopped)
		_, err := group.Do(
			t.Context(),
			"profile",
			func(context.Context) (string, error) { t.Error("must not start after close"); return "", nil },
		)
		require.ErrorIs(t, err, context.Canceled)
	})
}
