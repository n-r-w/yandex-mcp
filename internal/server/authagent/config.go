package authagent

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Config contains workstation auth-agent settings.
type Config struct {
	Port      int
	YCPath    string
	SSHTarget string
}

type envConfig struct {
	Port      int    `env:"YANDEX_MCP_AUTH_AGENT_PORT"              envDefault:"18765"`
	YCPath    string `env:"YANDEX_MCP_AUTH_AGENT_YC_PATH"           envDefault:"yc"`
	SSHTarget string `env:"YANDEX_MCP_SSH_TARGET,required,notEmpty"`
}

// LoadConfig reads the SSH destination, loopback port, and optional yc command override.
func LoadConfig() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse auth-agent config: %w", err)
	}
	if _, err := config.AuthAgentAddress(ec.Port); err != nil {
		return nil, err
	}
	return &Config{Port: ec.Port, YCPath: ec.YCPath, SSHTarget: ec.SSHTarget}, nil
}
