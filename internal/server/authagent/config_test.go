package authagent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Auth-agent needs no organization ID. Only native executable, loopback address,
// and explicit profile allowlist are accepted. Tests use a real executable path.
func TestLoadConfig(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_LISTEN", "127.0.0.1:18765")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_YC_PATH", executable)
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PROFILES", "work,personal")
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"work", "personal"}, cfg.Profiles)
	require.Equal(t, executable, cfg.YCPath)
	t.Run("missing executable diagnostic", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing-yc")
		t.Setenv("YANDEX_MCP_AUTH_AGENT_YC_PATH", missing)
		_, loadErr := LoadConfig()
		require.ErrorContains(t, loadErr, missing)
	})
	for _, tt := range []struct{ key, value string }{
		{"YANDEX_MCP_AUTH_AGENT_LISTEN", "0.0.0.0:18765"},
		{"YANDEX_MCP_AUTH_AGENT_YC_PATH", "yc"},
		{"YANDEX_MCP_AUTH_AGENT_PROFILES", ""},
		{"YANDEX_MCP_AUTH_AGENT_PROFILES", "work,,other"},
	} {
		t.Run(
			tt.key+tt.value,
			func(t *testing.T) { t.Setenv(tt.key, tt.value); _, err := LoadConfig(); require.Error(t, err) },
		)
	}
}
