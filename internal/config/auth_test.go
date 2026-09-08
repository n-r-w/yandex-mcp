package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAuthenticationConfig covers the tool deadline and remote port/profile settings.
func TestAuthenticationConfig(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "local")
	t.Setenv("YANDEX_MCP_TOOL_TIMEOUT", "300")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 300*time.Second, cfg.ToolTimeout)
	require.Equal(t, 18765, cfg.AuthAgentPort)
	tests := []struct {
		name, source, port, profile, timeout string
		valid                                bool
	}{
		{"remote", "remote", "28765", "work", "300", true},
		{"source", "other", "18765", "", "300", false},
		{"profile", "remote", "18765", "", "300", false},
		{"zero port", "remote", "0", "work", "300", false},
		{"negative port", "remote", "-1", "work", "300", false},
		{"large port", "remote", "65536", "work", "300", false},
		{"invalid port", "remote", "word", "work", "300", false},
		{"zero timeout", "local", "18765", "", "0", false},
		{"negative timeout", "local", "18765", "", "-1", false},
		{"overflow timeout", "local", "18765", "", "9223372036854775807", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("YANDEX_MCP_TOKEN_SOURCE", tt.source)
			t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", tt.port)
			t.Setenv("YANDEX_CLI_PROFILE", tt.profile)
			t.Setenv("YANDEX_MCP_TOOL_TIMEOUT", tt.timeout)
			cfg, err := Load()
			if tt.valid {
				require.NoError(t, err)
				require.Equal(t, 28765, cfg.AuthAgentPort)
			} else {
				require.Error(t, err)
			}
		})
	}
}
