// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
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

	// AttachAllowedExtensions is the list of allowed attachment extensions (without dots).
	AttachAllowedExtensions []string

	// AttachAllowedDirs is the list of absolute directories allowed for saving attachments.
	AttachAllowedDirs []string
}

// envConfig is an intermediate struct for parsing environment variables.
type envConfig struct {
	WikiBaseURL        string `env:"YANDEX_WIKI_BASE_URL"`
	TrackerBaseURL     string `env:"YANDEX_TRACKER_BASE_URL"`
	CloudOrgID         string `env:"YANDEX_CLOUD_ORG_ID,required"`
	RefreshPeriodHours int    `env:"YANDEX_IAM_TOKEN_REFRESH_PERIOD" envDefault:"10"`
	HTTPTimeoutSeconds int    `env:"YANDEX_HTTP_TIMEOUT" envDefault:"30"`
	AttachExtensions   string `env:"YANDEX_MCP_ATTACH_EXT"`
	AttachDirs         string `env:"YANDEX_MCP_ATTACH_DIR"`
}

// Load parses configuration from environment variables and validates it.
// It should be called once at application startup.
func Load() (*Config, error) {
	var ec envConfig
	if err := env.Parse(&ec); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	allowedExtensions, err := parseExtensionEnv(ec.AttachExtensions, "YANDEX_MCP_ATTACH_EXT")
	if err != nil {
		return nil, err
	}
	if len(allowedExtensions) == 0 {
		allowedExtensions = defaultAttachExtensions()
	}

	allowedDirs, err := parseDirEnv(ec.AttachDirs, "YANDEX_MCP_ATTACH_DIR")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		WikiBaseURL:             applyDefault(ec.WikiBaseURL, defaultWikiBaseURL),
		TrackerBaseURL:          applyDefault(ec.TrackerBaseURL, defaultTrackerBaseURL),
		CloudOrgID:              ec.CloudOrgID,
		IAMTokenRefreshPeriod:   resolveRefreshPeriod(ec.RefreshPeriodHours),
		HTTPTimeout:             time.Duration(ec.HTTPTimeoutSeconds) * time.Second,
		AttachAllowedExtensions: allowedExtensions,
		AttachAllowedDirs:       allowedDirs,
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

// defaultAttachExtensions provides the default allowlist for attachment saving.
func defaultAttachExtensions() []string {
	return []string{
		"txt",
		"json",
		"jsonc",
		"yaml",
		"yml",
		"md",
		"pdf",
		"doc",
		"docx",
		"rtf",
		"odt",
		"xls",
		"xlsx",
		"ods",
		"csv",
		"tsv",
		"ppt",
		"pptx",
		"odp",
		"jpg",
		"jpeg",
		"png",
		"tiff",
		"tif",
		"gif",
		"bmp",
		"webp",
		"zip",
		"7z",
		"tar",
		"tgz",
		"tar.gz",
		"gz",
		"bz2",
		"xz",
		"rar",
	}
}

// parseExtensionEnv normalizes extension settings for safe downstream checks.
func parseExtensionEnv(rawValue, envName string) ([]string, error) {
	items, err := parseCSV(rawValue, envName)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	normalized := make([]string, 0, len(items))
	for _, item := range items {
		ext := strings.ToLower(strings.TrimPrefix(item, "."))
		if ext == "" {
			return nil, fmt.Errorf("%s: empty extension", envName)
		}
		if !isValidExtension(ext) {
			return nil, fmt.Errorf("%s: invalid extension %q", envName, item)
		}
		normalized = append(normalized, ext)
	}

	return normalized, nil
}

// parseDirEnv validates directory overrides to prevent unsafe paths.
func parseDirEnv(rawValue, envName string) ([]string, error) {
	items, err := parseCSV(rawValue, envName)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	normalized := make([]string, 0, len(items))
	for _, item := range items {
		cleaned := filepath.Clean(item)
		if !filepath.IsAbs(cleaned) {
			return nil, fmt.Errorf("%s: must be absolute path, got %q", envName, item)
		}
		normalized = append(normalized, cleaned)
	}

	return normalized, nil
}

// parseCSV ensures consistent parsing for comma-delimited env values.
func parseCSV(rawValue, envName string) ([]string, error) {
	if strings.TrimSpace(rawValue) == "" {
		return nil, nil
	}
	parts := strings.Split(rawValue, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			return nil, fmt.Errorf("%s: empty value is not allowed", envName)
		}
		items = append(items, item)
	}
	return items, nil
}

// isValidExtension rejects unexpected characters in allowlists.
func isValidExtension(ext string) bool {
	for _, r := range ext {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.':
		default:
			return false
		}
	}
	return true
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
	if len(c.AttachAllowedExtensions) == 0 {
		errs = append(errs, errors.New("YANDEX_MCP_ATTACH_EXT resolved to an empty list"))
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
