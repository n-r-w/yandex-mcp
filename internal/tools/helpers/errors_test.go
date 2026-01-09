package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestToSafeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		serviceName  string
		wantContains string
	}{
		{
			name: "upstream error",
			err: domain.NewUpstreamError(
				domain.ServiceTracker,
				"search_issues",
				500,
				"",
				"server error",
				"",
			),
			serviceName:  "tracker",
			wantContains: "HTTP 500",
		},
		{
			name:         "decode response error",
			err:          errors.New("decode response: json: cannot unmarshal number into Go struct field Queue.id of type string"),
			serviceName:  "tracker",
			wantContains: "decode response:",
		},
		{
			name:         "read response body error",
			err:          errors.New("read response body: unexpected EOF"),
			serviceName:  "wiki",
			wantContains: "read response body:",
		},
		{
			name:         "parse base URL error",
			err:          errors.New("parse base URL: invalid URL"),
			serviceName:  "tracker",
			wantContains: "parse base URL:",
		},
		{
			name:         "create request error",
			err:          errors.New("create request: invalid method"),
			serviceName:  "wiki",
			wantContains: "create request:",
		},
		{
			name:         "marshal request body error",
			err:          errors.New("marshal request body: unsupported type"),
			serviceName:  "tracker",
			wantContains: "marshal request body:",
		},
		{
			name:         "execute request error",
			err:          errors.New("execute request: connection refused"),
			serviceName:  "tracker",
			wantContains: "execute request:",
		},
		{
			name:         "get token error",
			err:          errors.New("get token: token expired"),
			serviceName:  "tracker",
			wantContains: "get token:",
		},
		{
			name:         "unknown error",
			err:          errors.New("something went wrong"),
			serviceName:  "tracker",
			wantContains: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ToSafeError(context.Background(), tt.err, tt.serviceName)
			assert.Contains(t, result.Error(), tt.wantContains)
		})
	}
}
