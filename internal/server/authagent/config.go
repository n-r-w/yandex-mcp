package authagent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Config contains workstation auth-agent settings.
type Config struct {
	Port       int
	YCPath     string
	SSHTargets []string
}

type envConfig struct {
	Port       int      `env:"YANDEX_MCP_AUTH_AGENT_PORT"               envDefault:"18765"`
	YCPath     string   `env:"YANDEX_MCP_AUTH_AGENT_YC_PATH"            envDefault:"yc"`
	SSHTargets []string `env:"YANDEX_MCP_SSH_TARGETS,required,notEmpty"                    envSeparator:","`
}

// LoadConfig reads SSH destinations, loopback port, and the optional yc command override.
func LoadConfig() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse auth-agent config: %w", err)
	}
	if _, err := config.AuthAgentAddress(ec.Port); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(ec.SSHTargets))
	for index, target := range ec.SSHTargets {
		target = strings.TrimSpace(target)
		if target == "" {
			return nil, errors.New("YANDEX_MCP_SSH_TARGETS contains an empty target")
		}
		if _, exists := seen[target]; exists {
			return nil, fmt.Errorf("YANDEX_MCP_SSH_TARGETS contains duplicate target %q", target)
		}
		ec.SSHTargets[index] = target
		seen[target] = struct{}{}
	}
	return &Config{Port: ec.Port, YCPath: ec.YCPath, SSHTargets: ec.SSHTargets}, nil
}
