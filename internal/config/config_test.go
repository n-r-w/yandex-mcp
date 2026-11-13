package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_AllValuesProvided(t *testing.T) {
	t.Setenv("YANDEX_WIKI_BASE_URL", "https://wiki.example.com")
	t.Setenv("YANDEX_TRACKER_BASE_URL", "https://tracker.example.com")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "org-123")
	t.Setenv("YANDEX_CLI_PROFILE", "tracker-profile")
	t.Setenv("YANDEX_IAM_TOKEN_REFRESH_PERIOD", "5")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://wiki.example.com", cfg.WikiBaseURL)
	assert.Equal(t, "https://tracker.example.com", cfg.TrackerBaseURL)
	assert.Equal(t, "org-123", cfg.CloudOrgID)
	assert.Equal(t, "tracker-profile", cfg.CLIProfile)
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
	assert.Empty(t, cfg.CLIProfile)
	assert.Equal(t, defaultRefreshHours*time.Hour, cfg.IAMTokenRefreshPeriod)
}

func TestLoad_MissingCloudOrgID(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_CLOUD_ORG_ID")
}

// TestLoad_OAuthOrganization covers the Yandex 360 organization with a file-based OAuth token.
func TestLoad_OAuthOrganization(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")
	t.Setenv("YANDEX_ORG_ID", " 1234567 ")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "oauth")
	t.Setenv("YANDEX_OAUTH_TOKEN_FILE", tokenFile)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "1234567", cfg.OrgID)
	assert.Empty(t, cfg.CloudOrgID)
	assert.Equal(t, tokenFile, cfg.OAuthTokenFile)
	headerName, headerValue := cfg.OrgHeader()
	assert.Equal(t, "X-Org-Id", headerName)
	assert.Equal(t, "1234567", headerValue)
	assert.Equal(t, "OAuth", cfg.AuthScheme())
}

// TestLoad_CloudOrganizationHeaders keeps IAM authentication on the Cloud organization header.
func TestLoad_CloudOrganizationHeaders(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "cloud-org")

	cfg, err := Load()

	require.NoError(t, err)
	headerName, headerValue := cfg.OrgHeader()
	assert.Equal(t, "X-Cloud-Org-Id", headerName)
	assert.Equal(t, "cloud-org", headerValue)
	assert.Equal(t, "Bearer", cfg.AuthScheme())
}

// TestLoad_OAuthWithCloudOrganization allows an OAuth token for a Cloud organization.
func TestLoad_OAuthWithCloudOrganization(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "cloud-org")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "oauth")
	t.Setenv("YANDEX_OAUTH_TOKEN_FILE", filepath.Join(t.TempDir(), "token"))

	cfg, err := Load()

	require.NoError(t, err)
	headerName, _ := cfg.OrgHeader()
	assert.Equal(t, "X-Cloud-Org-Id", headerName)
	assert.Equal(t, "OAuth", cfg.AuthScheme())
}

// TestLoad_OrganizationAndOAuthValidation rejects ambiguous or unusable authentication settings.
func TestLoad_OrganizationAndOAuthValidation(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	tests := []struct {
		name, cloudOrgID, orgID, source, tokenFile, want string
	}{
		{"both organizations", "cloud-org", "1234567", "oauth", tokenFile, "mutually exclusive"},
		{"360 organization with IAM", "", "1234567", "local", "", "YANDEX_ORG_ID requires"},
		{"missing token file", "", "1234567", "oauth", "", "YANDEX_OAUTH_TOKEN_FILE"},
		{"relative token file", "", "1234567", "oauth", "token", "YANDEX_OAUTH_TOKEN_FILE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("YANDEX_CLOUD_ORG_ID", tt.cloudOrgID)
			t.Setenv("YANDEX_ORG_ID", tt.orgID)
			t.Setenv("YANDEX_MCP_TOKEN_SOURCE", tt.source)
			t.Setenv("YANDEX_OAUTH_TOKEN_FILE", tt.tokenFile)

			cfg, err := Load()

			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
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
	assert.Contains(t, cfg.AttachAllowedExtensions, "html")
	assert.Contains(t, cfg.AttachAllowedExtensions, "htm")
	assert.Equal(t, defaultTextAttachExtensions(), cfg.AttachViewExtensions)
	assert.NotContains(t, cfg.AttachViewExtensions, "html")
	assert.NotContains(t, cfg.AttachViewExtensions, "htm")
	assert.Nil(t, cfg.AttachAllowedDirs)
	assert.Equal(t, int64(defaultAttachInlineMaxBytes), cfg.AttachInlineMaxBytes)
}

func TestLoad_AttachInlineMaxBytesOverride(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_INLINE_MAX_BYTES", "2048")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, int64(2048), cfg.AttachInlineMaxBytes)
}

func TestLoad_AttachInlineMaxBytesInvalid(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_MCP_ATTACH_INLINE_MAX_BYTES", "0")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_MCP_ATTACH_INLINE_MAX_BYTES")
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
	dirOne := t.TempDir()
	dirTwo := t.TempDir()
	t.Setenv("YANDEX_MCP_ATTACH_DIR", dirOne+","+dirTwo)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, []string{dirOne, dirTwo}, cfg.AttachAllowedDirs)
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
