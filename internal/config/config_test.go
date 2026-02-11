package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_AllValuesProvided(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "https://wiki.example.com")
	t.Setenv("YANDEX_TRACKER_BASE_URL", "https://tracker.example.com")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")
	t.Setenv("YANDEX_IAM_TOKEN_REFRESH_PERIOD", "5")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://wiki.example.com", cfg.WikiBaseURL)
	assert.Equal(t, "https://tracker.example.com", cfg.TrackerBaseURL)
	assert.Equal(t, "org-123", cfg.CloudOrgID)
	assert.Equal(t, 5*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_TrackerBaseURLUsesDefault(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "https://wiki.example.com")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultTrackerBaseURL, cfg.TrackerBaseURL)
	assert.Equal(t, defaultRefreshHours*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_WikiBaseURLIsOptional(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-456")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultWikiBaseURL, cfg.WikiBaseURL)
	assert.Equal(t, defaultTrackerBaseURL, cfg.TrackerBaseURL)
	assert.Equal(t, "org-456", cfg.CloudOrgID)
}

func TestLoad_DefaultRefreshPeriod(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultRefreshHours*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_MissingCloudOrgID(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_CLOUD_ORG_ID")
}

func TestLoad_WikiBaseURLNotHTTPS(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "http://wiki.example.com")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "must use https scheme")
}

func TestLoad_TrackerBaseURLNotHTTPS(t *testing.T) {
	t.Setenv("YANDEX_TRACKER_BASE_URL", "http://tracker.example.com")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "must use https scheme")
}

func TestLoad_WikiBaseURLMissingHost(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "https://")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "missing host")
}

func TestLoad_TrackerBaseURLMissingHost(t *testing.T) {
	t.Setenv("YANDEX_TRACKER_BASE_URL", "https://")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "missing host")
}

func TestLoad_NegativeRefreshPeriodUsesDefault(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_IAM_TOKEN_REFRESH_PERIOD", "-5")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultRefreshHours*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_ZeroRefreshPeriodUsesDefault(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_IAM_TOKEN_REFRESH_PERIOD", "0")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultRefreshHours*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_DefaultTrackerURL(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://api.tracker.yandex.net", cfg.TrackerBaseURL)
}

func TestLoad_DefaultAttachExtensions(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, defaultAttachExtensions(), cfg.AttachAllowedExtensions)
	assert.Equal(t, defaultTextAttachExtensions(), cfg.AttachViewExtensions)
	assert.Nil(t, cfg.AttachAllowedDirs)
}

func TestLoad_AttachExtensionsOverride(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_EXT", "txt,md,tar.gz")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"txt", "md", "tar.gz"}, cfg.AttachAllowedExtensions)
}

func TestLoad_AttachViewExtensionsOverride(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_VIEW_EXT", "txt,md,rtf")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"txt", "md", "rtf"}, cfg.AttachViewExtensions)
}

func TestLoad_AttachDirsOverride(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_DIR", "/tmp,/var/tmp")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"/tmp", "/var/tmp"}, cfg.AttachAllowedDirs)
}

func TestLoad_AttachDirsMustBeAbsolute(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_DIR", "relative")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_MCP_ATTACH_DIR")
}

func TestLoad_AttachExtensionsInvalid(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_EXT", "t*xt")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_MCP_ATTACH_EXT")
}

func TestLoad_AttachViewExtensionsInvalid(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_VIEW_EXT", "t*xt")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_MCP_ATTACH_VIEW_EXT")
}
