package authagent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Config contains workstation-only settings.
type Config struct {
	ListenAddress string
	YCPath        string
	Profiles      []string
}

type envConfig struct {
	ListenAddress string   `env:"YANDEX_MCP_AUTH_AGENT_LISTEN"            envDefault:"127.0.0.1:18765"`
	YCPath        string   `env:"YANDEX_MCP_AUTH_AGENT_YC_PATH,required"`
	Profiles      []string `env:"YANDEX_MCP_AUTH_AGENT_PROFILES,required"                              envSeparator:","`
}

// LoadConfig reads workstation settings, independently of MCP API settings.
func LoadConfig() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse auth-agent config: %w", err)
	}
	if err := config.ValidateLoopbackAddress(ec.ListenAddress); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(ec.YCPath) {
		return nil, errors.New("YANDEX_MCP_AUTH_AGENT_YC_PATH must be an absolute native executable path")
	}
	info, err := os.Stat(ec.YCPath)
	if err != nil {
		return nil, fmt.Errorf("YANDEX_MCP_AUTH_AGENT_YC_PATH: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("YANDEX_MCP_AUTH_AGENT_YC_PATH must identify a regular executable file")
	}
	if len(ec.Profiles) == 0 {
		return nil, errors.New("YANDEX_MCP_AUTH_AGENT_PROFILES must contain allowed profiles")
	}
	for i, profile := range ec.Profiles {
		ec.Profiles[i] = strings.TrimSpace(profile)
		if ec.Profiles[i] == "" {
			return nil, errors.New("YANDEX_MCP_AUTH_AGENT_PROFILES contains an empty profile")
		}
	}
	return &Config{ListenAddress: ec.ListenAddress, YCPath: ec.YCPath, Profiles: ec.Profiles}, nil
}
