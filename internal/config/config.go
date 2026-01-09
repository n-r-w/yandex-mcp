// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	defaultTrackerBaseURL = "https://api.tracker.yandex.net"
	defaultWikiBaseURL    = "https://api.wiki.yandex.net"
	defaultRefreshHours   = 10
)

// Config holds static application configuration loaded from environment variables.
type Config struct {
	// WikiBaseURL is the base URL for Yandex Wiki API.
	WikiBaseURL string

	// TrackerBaseURL is the base URL for Yandex Tracker API.
	TrackerBaseURL string

	// CloudOrgID is the Yandex Cloud Organization ID for X-Cloud-Org-Id header.
	CloudOrgID string

	// IAMTokenRefreshPeriod is the period after which the IAM token should be refreshed.
	IAMTokenRefreshPeriod time.Duration

	// HTTPTimeout is the timeout for HTTP requests to Yandex APIs.
	HTTPTimeout time.Duration
}

// envConfig is an intermediate struct for parsing environment variables.
type envConfig struct {
	WikiBaseURL        string `env:"YANDEX_WIKI_BASE_URL"`
	TrackerBaseURL     string `env:"YANDEX_TRACKER_BASE_URL"`
	CloudOrgID         string `env:"YANDEX_CLOUD_ORG_ID,required"`
	RefreshPeriodHours int    `env:"YANDEX_IAM_TOKEN_REFRESH_PERIOD" envDefault:"10"`
	HTTPTimeoutSeconds int    `env:"YANDEX_HTTP_TIMEOUT" envDefault:"30"`
}

// Load parses configuration from environment variables and validates it.
// It should be called once at application startup.
func Load() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	cfg := &Config{
		WikiBaseURL:           applyDefault(ec.WikiBaseURL, defaultWikiBaseURL),
		TrackerBaseURL:        applyDefault(ec.TrackerBaseURL, defaultTrackerBaseURL),
		CloudOrgID:            ec.CloudOrgID,
		IAMTokenRefreshPeriod: resolveRefreshPeriod(ec.RefreshPeriodHours),
		HTTPTimeout:           time.Duration(ec.HTTPTimeoutSeconds) * time.Second,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func applyDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func resolveRefreshPeriod(hours int) time.Duration {
	if hours <= 0 {
		return defaultRefreshHours * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func (c *Config) validate() error {
	var errs []error

	if err := validateHTTPSURL(c.WikiBaseURL, "YANDEX_WIKI_BASE_URL"); err != nil {
		errs = append(errs, err)
	}

	if err := validateHTTPSURL(c.TrackerBaseURL, "YANDEX_TRACKER_BASE_URL"); err != nil {
		errs = append(errs, err)
	}

	if c.CloudOrgID == "" {
		errs = append(errs, errors.New("YANDEX_CLOUD_ORG_ID is required"))
	}

	return errors.Join(errs...)
}

func validateHTTPSURL(rawURL, envName string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%s: invalid URL: %w", envName, err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("%s: must use https scheme, got %q", envName, parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("%s: missing host", envName)
	}

	return nil
}
