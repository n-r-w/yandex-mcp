package authagent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Config contains workstation auth-agent settings.
type Config struct {
	Port      int
	YCPath    string
	SSHTarget string
	// OAuthDir holds OAuth token files named after profiles. Empty when no directory can be resolved.
	OAuthDir string
}

type envConfig struct {
	Port      int    `env:"YANDEX_MCP_AUTH_AGENT_PORT"              envDefault:"18765"`
	YCPath    string `env:"YANDEX_MCP_AUTH_AGENT_YC_PATH"           envDefault:"yc"`
	OAuthDir  string `env:"YANDEX_MCP_AUTH_AGENT_OAUTH_DIR"`
	SSHTarget string `env:"YANDEX_MCP_SSH_TARGET,required,notEmpty"`
}

// LoadConfig reads the SSH destination, loopback port, optional yc command override,
// and the directory of OAuth profile files.
func LoadConfig() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse auth-agent config: %w", err)
	}
	if _, err := config.AuthAgentAddress(ec.Port); err != nil {
		return nil, err
	}
	oauthDir, err := resolveOAuthDir(ec.OAuthDir)
	if err != nil {
		return nil, err
	}
	return &Config{Port: ec.Port, YCPath: ec.YCPath, SSHTarget: ec.SSHTarget, OAuthDir: oauthDir}, nil
}

// resolveOAuthDir returns the explicit directory or the per-user default.
// A workstation without a user configuration directory still serves IAM tokens.
func resolveOAuthDir(explicit string) (string, error) {
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return "", errors.New("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR must be an absolute path")
		}
		return filepath.Clean(explicit), nil
	}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		//nolint:nilerr // OAuth requests then fail explicitly in the token source
		return "", nil
	}
	return filepath.Join(userConfigDir, oauthDirName, oauthSubdirName), nil
}
