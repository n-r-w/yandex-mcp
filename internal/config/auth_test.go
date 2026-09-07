package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Environment-only configuration: local defaults, explicit remote profile,
// positive overall timeout and strict loopback URL. No external dependencies.
func TestAuthenticationConfig(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "local")
	t.Setenv("YANDEX_MCP_TOOL_TIMEOUT", "300")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 300*time.Second, cfg.ToolTimeout)
	tests := []struct {
		name, source, endpoint, profile, timeout string
		valid                                    bool
	}{
		{"remote", "remote", "http://127.0.0.1:18765", "work", "300", true},
		{"source", "other", "", "", "300", false},
		{"profile", "remote", "http://127.0.0.1:18765", "", "300", false},
		{"endpoint", "remote", "http://example.com:18765", "work", "300", false},
		{"zero", "local", "", "", "0", false},
		{"negative", "local", "", "", "-1", false},
		{"overflow", "local", "", "", "9223372036854775807", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("YANDEX_MCP_TOKEN_SOURCE", tt.source)
			t.Setenv("YANDEX_MCP_AUTH_AGENT_URL", tt.endpoint)
			t.Setenv("YANDEX_CLI_PROFILE", tt.profile)
			t.Setenv("YANDEX_MCP_TOOL_TIMEOUT", tt.timeout)
			_, err := Load()
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
