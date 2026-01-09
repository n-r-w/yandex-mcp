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
	defaultTrackerBaseURL       = "https://api.tracker.yandex.net"
	defaultWikiBaseURL          = "https://api.wiki.yandex.net"
	defaultIAMTokenRefreshHours = 10
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
}

// envConfig is an intermediate struct for parsing environment variables.
type envConfig struct {
	WikiBaseURL              string `env:"YANDEX_WIKI_BASE_URL"`
	TrackerBaseURL           string `env:"YANDEX_TRACKER_BASE_URL"`
	CloudOrgID               string `env:"YANDEX_CLOUD_ORG_ID,required"`
	IAMTokenRefreshPeriodHrs int    `env:"YANDEX_IAM_TOKEN_REFRESH_PERIOD" envDefault:"10"`
}

// Load parses configuration from environment variables and validates it.
// It should be called once at application startup.
func Load() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	cfg := &Config{
		WikiBaseURL:           ec.WikiBaseURL,
		TrackerBaseURL:        ec.TrackerBaseURL,
		CloudOrgID:            ec.CloudOrgID,
		IAMTokenRefreshPeriod: time.Duration(ec.IAMTokenRefreshPeriodHrs) * time.Hour,
	}

	if cfg.TrackerBaseURL == "" {
		cfg.TrackerBaseURL = defaultTrackerBaseURL
	}

	if cfg.WikiBaseURL == "" {
		cfg.WikiBaseURL = defaultWikiBaseURL
	}

	if cfg.IAMTokenRefreshPeriod <= 0 {
		cfg.IAMTokenRefreshPeriod = defaultIAMTokenRefreshHours * time.Hour
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var errs []error

	if c.WikiBaseURL != "" {
		if err := validateHTTPSURL(c.WikiBaseURL, "YANDEX_WIKI_BASE_URL"); err != nil {
			errs = append(errs, err)
		}
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
