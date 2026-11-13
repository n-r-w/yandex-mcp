// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	defaultTrackerBaseURL       = "https://api.tracker.yandex.net"
	defaultWikiBaseURL          = "https://api.wiki.yandex.net"
	defaultRefreshHours         = 10
	defaultAttachInlineMaxBytes = 10 * 1024 * 1024
)

// Config holds static application configuration loaded from environment variables.
type Config struct {
	// WikiBaseURL is the base URL for Yandex Wiki API.
	WikiBaseURL string

	// TrackerBaseURL is the base URL for Yandex Tracker API.
	TrackerBaseURL string

	// CloudOrgID is the Yandex Cloud Organization ID for X-Cloud-Org-Id header.
	CloudOrgID string

	// OrgID is the Yandex 360 for Business organization ID for X-Org-Id header.
	OrgID string

	// OAuthTokenFile is the absolute path to a file holding a Yandex OAuth token, used for local acquisition.
	OAuthTokenFile string

	// OAuthProfile names the OAuth token that auth-agent resolves to a file, used for remote acquisition.
	OAuthProfile string

	// TokenType selects the IAM or OAuth token type independently of the acquisition location.
	TokenType string

	// CLIProfile is the optional yc CLI profile used for IAM token retrieval.
	CLIProfile string

	// IAMTokenRefreshPeriod is the period after which the cached IAM or OAuth token is refreshed.
	IAMTokenRefreshPeriod time.Duration

	// ToolTimeout bounds the complete tools/call operation.
	ToolTimeout time.Duration

	// TokenSource selects local or remote acquisition.
	TokenSource string

	// AuthAgentPort is the server-side tunnel port used in remote mode.
	AuthAgentPort int

	// AttachAllowedExtensions is the list of allowed attachment extensions (without dots).
	AttachAllowedExtensions []string

	// AttachViewExtensions is the list of allowed attachment extensions for inline viewing (without dots).
	AttachViewExtensions []string

	// AttachAllowedDirs is the list of absolute directories allowed for saving attachments.
	AttachAllowedDirs []string

	// AttachInlineMaxBytes is the maximum size of attachment content returned inline.
	AttachInlineMaxBytes int64
}

// envConfig is an intermediate struct for parsing environment variables.
type envConfig struct {
	WikiBaseURL          string `env:"YANDEX_WIKI_BASE_URL"`
	TrackerBaseURL       string `env:"YANDEX_TRACKER_BASE_URL"`
	CloudOrgID           string `env:"YANDEX_CLOUD_ORG_ID"`
	OrgID                string `env:"YANDEX_ORG_ID"`
	OAuthTokenFile       string `env:"YANDEX_OAUTH_TOKEN_FILE"`
	OAuthProfile         string `env:"YANDEX_OAUTH_PROFILE"`
	TokenType            string `env:"YANDEX_MCP_TOKEN_TYPE"              envDefault:"iam"`
	CLIProfile           string `env:"YANDEX_CLI_PROFILE"`
	RefreshPeriodHours   int    `env:"YANDEX_IAM_TOKEN_REFRESH_PERIOD"    envDefault:"10"`
	ToolTimeoutSeconds   int    `env:"YANDEX_MCP_TOOL_TIMEOUT"            envDefault:"300"`
	TokenSource          string `env:"YANDEX_MCP_TOKEN_SOURCE"            envDefault:"local"`
	AuthAgentPort        int    `env:"YANDEX_MCP_AUTH_AGENT_PORT"         envDefault:"18765"`
	AttachExtensions     string `env:"YANDEX_MCP_ATTACH_EXT"`
	AttachViewExts       string `env:"YANDEX_MCP_ATTACH_VIEW_EXT"`
	AttachDirs           string `env:"YANDEX_MCP_ATTACH_DIR"`
	AttachInlineMaxBytes int64  `env:"YANDEX_MCP_ATTACH_INLINE_MAX_BYTES" envDefault:"10485760"`
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

	viewExtensions, err := parseExtensionEnv(ec.AttachViewExts, "YANDEX_MCP_ATTACH_VIEW_EXT")
	if err != nil {
		return nil, err
	}
	if len(viewExtensions) == 0 {
		viewExtensions = defaultTextAttachExtensions()
	}

	allowedDirs, err := parseDirEnv(ec.AttachDirs, "YANDEX_MCP_ATTACH_DIR")
	if err != nil {
		return nil, err
	}

	if ec.ToolTimeoutSeconds <= 0 || int64(ec.ToolTimeoutSeconds) > int64((1<<63-1)/time.Second) {
		return nil, errors.New("YANDEX_MCP_TOOL_TIMEOUT must be positive seconds within time.Duration range")
	}

	cfg := &Config{
		WikiBaseURL:             applyDefault(ec.WikiBaseURL, defaultWikiBaseURL),
		TrackerBaseURL:          applyDefault(ec.TrackerBaseURL, defaultTrackerBaseURL),
		CloudOrgID:              strings.TrimSpace(ec.CloudOrgID),
		OrgID:                   strings.TrimSpace(ec.OrgID),
		OAuthTokenFile:          strings.TrimSpace(ec.OAuthTokenFile),
		OAuthProfile:            strings.TrimSpace(ec.OAuthProfile),
		TokenType:               ec.TokenType,
		CLIProfile:              ec.CLIProfile,
		IAMTokenRefreshPeriod:   resolveRefreshPeriod(ec.RefreshPeriodHours),
		ToolTimeout:             time.Duration(ec.ToolTimeoutSeconds) * time.Second,
		TokenSource:             ec.TokenSource,
		AuthAgentPort:           ec.AuthAgentPort,
		AttachAllowedExtensions: allowedExtensions,
		AttachViewExtensions:    viewExtensions,
		AttachAllowedDirs:       allowedDirs,
		AttachInlineMaxBytes:    ec.AttachInlineMaxBytes,
	}

	if validateErr := cfg.validate(); validateErr != nil {
		return nil, fmt.Errorf("validate config: %w", validateErr)
	}

	return cfg, nil
}

func applyDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// defaultTextAttachExtensions provides the default allowlist for inline text viewing.
func defaultTextAttachExtensions() []string {
	return []string{
		"txt",
		"json",
		"jsonc",
		"yaml",
		"yml",
		"md",
		"csv",
		"tsv",
		"rtf",
	}
}

// defaultAttachExtensions provides the default allowlist for attachment saving.
func defaultAttachExtensions() []string {
	return append(defaultTextAttachExtensions(), defaultNonTextAttachExtensions()...)
}

// defaultNonTextAttachExtensions provides the default allowlist additions for non-text files.
func defaultNonTextAttachExtensions() []string {
	return []string{
		"html",
		"htm",
		"pdf",
		"doc",
		"docx",
		"odt",
		"xls",
		"xlsx",
		"ods",
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

	errs = append(errs, c.validateTokenAcquisition()...)

	if err := validateHTTPSURL(c.WikiBaseURL, "YANDEX_WIKI_BASE_URL"); err != nil {
		errs = append(errs, err)
	}

	if err := validateHTTPSURL(c.TrackerBaseURL, "YANDEX_TRACKER_BASE_URL"); err != nil {
		errs = append(errs, err)
	}

	if err := c.validateOrganization(); err != nil {
		errs = append(errs, err)
	}
	if len(c.AttachAllowedExtensions) == 0 {
		errs = append(errs, errors.New("YANDEX_MCP_ATTACH_EXT resolved to an empty list"))
	}
	if len(c.AttachViewExtensions) == 0 {
		errs = append(errs, errors.New("YANDEX_MCP_ATTACH_VIEW_EXT resolved to an empty list"))
	}
	if c.AttachInlineMaxBytes <= 0 {
		errs = append(errs, errors.New("YANDEX_MCP_ATTACH_INLINE_MAX_BYTES must be positive"))
	}

	return errors.Join(errs...)
}

// validateTokenAcquisition checks the acquisition location and the token type independently,
// then the settings required by their combination.
func (c *Config) validateTokenAcquisition() []error {
	var errs []error

	if c.TokenSource != TokenSourceLocal && c.TokenSource != TokenSourceRemote {
		errs = append(errs, errors.New("YANDEX_MCP_TOKEN_SOURCE must be local or remote"))
	}
	if c.TokenType != TokenTypeIAM && c.TokenType != TokenTypeOAuth {
		errs = append(errs, errors.New("YANDEX_MCP_TOKEN_TYPE must be iam or oauth"))
	}

	if c.TokenSource == TokenSourceRemote {
		if _, err := AuthAgentAddress(c.AuthAgentPort); err != nil {
			errs = append(errs, err)
		}
	}
	if err := c.validateTokenIdentity(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// validateTokenIdentity requires the setting that identifies the token for the selected
// combination: a profile name for remote acquisition, a file path for a local OAuth token.
func (c *Config) validateTokenIdentity() error {
	remote := c.TokenSource == TokenSourceRemote
	switch {
	case remote && c.TokenType == TokenTypeOAuth && c.OAuthProfile == "":
		return errors.New("YANDEX_OAUTH_PROFILE is required for remote OAuth tokens")
	case remote && c.TokenType == TokenTypeIAM && strings.TrimSpace(c.CLIProfile) == "":
		return errors.New("YANDEX_CLI_PROFILE is required in remote mode")
	case !remote && c.TokenType == TokenTypeOAuth && !filepath.IsAbs(c.OAuthTokenFile):
		return errors.New("YANDEX_OAUTH_TOKEN_FILE must be an absolute path for local OAuth tokens")
	}
	return nil
}

// Profile returns the profile name that identifies the token for the configured token type.
func (c *Config) Profile() string {
	if c.TokenType == TokenTypeOAuth {
		return c.OAuthProfile
	}
	return c.CLIProfile
}

// validateOrganization requires one organization type that the token type can authenticate.
func (c *Config) validateOrganization() error {
	switch {
	case c.CloudOrgID == "" && c.OrgID == "":
		return errors.New("YANDEX_CLOUD_ORG_ID or YANDEX_ORG_ID is required")
	case c.CloudOrgID != "" && c.OrgID != "":
		return errors.New("YANDEX_CLOUD_ORG_ID and YANDEX_ORG_ID are mutually exclusive")
	case c.OrgID != "" && c.TokenType != TokenTypeOAuth:
		// IAM tokens are accepted only by Yandex Cloud organizations.
		return errors.New("YANDEX_ORG_ID requires YANDEX_MCP_TOKEN_TYPE=oauth")
	}
	return nil
}

// OrgHeader returns the organization header name and value for the configured organization type.
func (c *Config) OrgHeader() (name, value string) {
	if c.OrgID != "" {
		return headerOrgID, c.OrgID
	}
	return headerCloudOrgID, c.CloudOrgID
}

// AuthScheme returns the Authorization header scheme for the configured token type.
func (c *Config) AuthScheme() string {
	if c.TokenType == TokenTypeOAuth {
		return authSchemeOAuth
	}
	return authSchemeBearer
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

// AuthAgentAddress fixes token exchange to IPv4 loopback on both machines.
func AuthAgentAddress(port int) (string, error) {
	if port < 1 || port > 65535 {
		return "", errors.New("YANDEX_MCP_AUTH_AGENT_PORT must be from 1 to 65535")
	}
	return "127.0.0.1:" + strconv.Itoa(port), nil
}
