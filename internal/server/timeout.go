package server

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func toolTimeout(timeout time.Duration) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(parent context.Context, method string, request mcp.Request) (mcp.Result, error) {
			if method != "tools/call" {
				return next(parent, method, request)
			}
			if err := parent.Err(); err != nil {
				return nil, err
			}
			ctx, cancel := context.WithTimeout(parent, timeout)
			defer cancel()
			completed := make(chan methodResult, 1)
			go func() { result, err := next(ctx, method, request); completed <- methodResult{result: result, err: err} }()
			var result methodResult
			select {
			case result = <-completed:
			case <-ctx.Done():
			}
			if err := parent.Err(); err != nil {
				return nil, err
			}
			if ctx.Err() != nil {
				//nolint:exhaustruct_v5 // only error content is returned
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: "tool call timed out"}},
				}, nil
			}
			return result.result, result.err
		}
	}
}
