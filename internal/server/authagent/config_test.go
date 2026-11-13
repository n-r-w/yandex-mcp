package authagent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadConfig covers multiple SSH targets, PATH-based defaults, explicit overrides, and invalid input.
func TestLoadConfig(t *testing.T) {
	t.Setenv("YANDEX_MCP_SSH_TARGETS", "first-server,second-server")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_YC_PATH", "")
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "yc", cfg.YCPath)
	require.Equal(t, 18765, cfg.Port)
	require.Equal(t, []string{"first-server", "second-server"}, cfg.SSHTargets)
	for name, targets := range map[string]string{
		"missing SSH targets":   "",
		"empty SSH target":      "first-server,,second-server",
		"duplicate SSH targets": "first-server,first-server",
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("YANDEX_MCP_SSH_TARGETS", targets)
			_, err := LoadConfig()
			require.ErrorContains(t, err, "YANDEX_MCP_SSH_TARGETS")
		})
	}
	executable, err := os.Executable()
	require.NoError(t, err)
	t.Setenv("YANDEX_MCP_AUTH_AGENT_YC_PATH", executable)
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "28765")
	cfg, err = LoadConfig()
	require.NoError(t, err)
	require.Equal(t, executable, cfg.YCPath)
	require.Equal(t, 28765, cfg.Port)
	for _, port := range []string{"0", "-1", "65536", "word"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", port)
			_, err := LoadConfig()
			require.Error(t, err)
			if port == "word" {
				require.ErrorContains(t, err, "word")
			} else {
				require.ErrorContains(t, err, "YANDEX_MCP_AUTH_AGENT_PORT")
			}
		})
	}
}

// TestLoadConfigOAuthDir covers the per-user default directory of OAuth profile files,
// an explicit override, and rejection of a relative override.
func TestLoadConfigOAuthDir(t *testing.T) {
	t.Setenv("YANDEX_MCP_SSH_TARGETS", "test-server")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "")
	t.Setenv("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR", "")
	userConfigDir, err := os.UserConfigDir()
	require.NoError(t, err)
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(userConfigDir, "yandex-mcp", "oauth"), cfg.OAuthDir)

	explicit := t.TempDir()
	t.Setenv("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR", explicit)
	cfg, err = LoadConfig()
	require.NoError(t, err)
	require.Equal(t, explicit, cfg.OAuthDir)

	t.Setenv("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR", "relative")
	_, err = LoadConfig()
	require.ErrorContains(t, err, "YANDEX_MCP_AUTH_AGENT_OAUTH_DIR")
}
