package server

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// Middleware covers tools/call before validation and returns a tool error at its
// own deadline even if a handler finishes later. Fake time removes scheduling delays.
func TestOverallToolTimeout(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		deadlineSet := make(chan bool, 1)
		finished := make(chan struct{})
		handler := toolTimeout(time.Second)(func(ctx context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			_, ok := ctx.Deadline()
			deadlineSet <- ok
			time.Sleep(2 * time.Second)
			close(finished)
			return &mcp.CallToolResult{}, nil
		})
		result, err := handler(t.Context(), "tools/call", nil)
		require.NoError(t, err)
		require.True(t, <-deadlineSet)
		callResult, ok := result.(*mcp.CallToolResult)
		require.True(t, ok)
		require.True(t, callResult.IsError)
		require.Equal(t, time.Second, time.Since(start))
		<-finished
	})
}

// Parent cancellation stays cancellation, while non-tool methods retain their
// original context and get no tool timeout. No transport dependencies are needed.
func TestTimeoutContextBoundaries(t *testing.T) {
	t.Parallel()
	handler := toolTimeout(
		time.Second,
	)(
		func(ctx context.Context, _ string, _ mcp.Request) (mcp.Result, error) { return nil, ctx.Err() },
	)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := handler(ctx, "tools/call", nil)
	require.ErrorIs(t, err, context.Canceled)
	handler = toolTimeout(time.Second)(func(ctx context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		_, ok := ctx.Deadline()
		require.False(t, ok)
		return &mcp.CallToolResult{}, nil
	})
	_, err = handler(t.Context(), "tools/list", nil)
	require.NoError(t, err)
}
