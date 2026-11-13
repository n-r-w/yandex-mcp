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

// acquisitionEnv describes token acquisition and organization settings of one configuration.
type acquisitionEnv struct {
	source, tokenType        string
	cliProfile, oauthProfile string
	tokenFile                string
	cloudOrgID, orgID        string
}

// variables lists every setting with its value, so a test also clears the variables a case leaves empty.
func (e acquisitionEnv) variables() map[string]string {
	return map[string]string{
		"YANDEX_MCP_TOKEN_SOURCE": e.source,
		"YANDEX_MCP_TOKEN_TYPE":   e.tokenType,
		"YANDEX_CLI_PROFILE":      e.cliProfile,
		"YANDEX_OAUTH_PROFILE":    e.oauthProfile,
		"YANDEX_OAUTH_TOKEN_FILE": e.tokenFile,
		"YANDEX_CLOUD_ORG_ID":     e.cloudOrgID,
		"YANDEX_ORG_ID":           e.orgID,
	}
}

// TestLoad_OAuthOrganization covers the Yandex 360 organization with a local OAuth token file.
func TestLoad_OAuthOrganization(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")
	t.Setenv("YANDEX_ORG_ID", " 1234567 ")
	t.Setenv("YANDEX_MCP_TOKEN_TYPE", "oauth")
	t.Setenv("YANDEX_OAUTH_TOKEN_FILE", tokenFile)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, TokenSourceLocal, cfg.TokenSource)
	assert.Equal(t, TokenTypeOAuth, cfg.TokenType)
	assert.Equal(t, "1234567", cfg.OrgID)
	assert.Empty(t, cfg.CloudOrgID)
	assert.Equal(t, tokenFile, cfg.OAuthTokenFile)
	headerName, headerValue := cfg.OrgHeader()
	assert.Equal(t, "X-Org-Id", headerName)
	assert.Equal(t, "1234567", headerValue)
	assert.Equal(t, "OAuth", cfg.AuthScheme())
}

// TestLoad_CloudOrganizationHeaders keeps the IAM token type and the Cloud organization header by default.
func TestLoad_CloudOrganizationHeaders(t *testing.T) {
	t.Setenv("YANDEX_CLOUD_ORG_ID", "cloud-org")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, TokenSourceLocal, cfg.TokenSource)
	assert.Equal(t, TokenTypeIAM, cfg.TokenType)
	headerName, headerValue := cfg.OrgHeader()
	assert.Equal(t, "X-Cloud-Org-Id", headerName)
	assert.Equal(t, "cloud-org", headerValue)
	assert.Equal(t, "Bearer", cfg.AuthScheme())
}

// TestLoad_TokenTypeIndependentOfSource verifies that the Authorization scheme, the profile and the
// organization compatibility follow the token type for both acquisition locations.
func TestLoad_TokenTypeIndependentOfSource(t *testing.T) {
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "")
	tokenFile := filepath.Join(t.TempDir(), "token")
	tests := []struct {
		name                                string
		env                                 acquisitionEnv
		wantScheme, wantProfile, wantHeader string
	}{
		{
			name: "local iam",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "iam",
				cliProfile:   "work",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			wantScheme:  "Bearer",
			wantProfile: "work",
			wantHeader:  "X-Cloud-Org-Id",
		},
		{
			name: "remote iam",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "iam",
				cliProfile:   "work",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			wantScheme:  "Bearer",
			wantProfile: "work",
			wantHeader:  "X-Cloud-Org-Id",
		},
		{
			name: "local oauth",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "oauth",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    tokenFile,
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			wantScheme:  "OAuth",
			wantProfile: "",
			wantHeader:  "X-Org-Id",
		},
		{
			name: "remote oauth",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "oauth",
				cliProfile:   "",
				oauthProfile: "work",
				tokenFile:    "",
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			wantScheme:  "OAuth",
			wantProfile: "work",
			wantHeader:  "X-Org-Id",
		},
		{
			name: "remote oauth with cloud organization",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "oauth",
				cliProfile:   "iam-profile",
				oauthProfile: "work",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			wantScheme:  "OAuth",
			wantProfile: "work",
			wantHeader:  "X-Cloud-Org-Id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, value := range tt.env.variables() {
				t.Setenv(name, value)
			}

			cfg, err := Load()

			require.NoError(t, err)
			assert.Equal(t, tt.env.source, cfg.TokenSource)
			assert.Equal(t, tt.env.tokenType, cfg.TokenType)
			assert.Equal(t, tt.wantScheme, cfg.AuthScheme())
			assert.Equal(t, tt.wantProfile, cfg.Profile())
			headerName, _ := cfg.OrgHeader()
			assert.Equal(t, tt.wantHeader, headerName)
		})
	}
}

// TestLoad_TokenAcquisitionValidation rejects ambiguous or unusable authentication settings.
func TestLoad_TokenAcquisitionValidation(t *testing.T) {
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", "")
	tokenFile := filepath.Join(t.TempDir(), "token")
	tests := []struct {
		name string
		env  acquisitionEnv
		want string
	}{
		{
			name: "both organizations",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "oauth",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    tokenFile,
				cloudOrgID:   "cloud-org",
				orgID:        "1234567",
			},
			want: "mutually exclusive",
		},
		{
			name: "360 organization with local iam",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "iam",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			want: "YANDEX_ORG_ID requires YANDEX_MCP_TOKEN_TYPE=oauth",
		},
		{
			name: "360 organization with remote iam",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "iam",
				cliProfile:   "work",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			want: "YANDEX_ORG_ID requires YANDEX_MCP_TOKEN_TYPE=oauth",
		},
		{
			name: "oauth is not an acquisition location",
			env: acquisitionEnv{
				source:       "oauth",
				tokenType:    "iam",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			want: "YANDEX_MCP_TOKEN_SOURCE must be local or remote",
		},
		{
			name: "unsupported token type",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "bearer",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			want: "YANDEX_MCP_TOKEN_TYPE must be iam or oauth",
		},
		{
			name: "local oauth without token file",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "oauth",
				cliProfile:   "",
				oauthProfile: "work",
				tokenFile:    "",
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			want: "YANDEX_OAUTH_TOKEN_FILE",
		},
		{
			name: "local oauth with relative token file",
			env: acquisitionEnv{
				source:       "local",
				tokenType:    "oauth",
				cliProfile:   "",
				oauthProfile: "",
				tokenFile:    "token",
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			want: "YANDEX_OAUTH_TOKEN_FILE",
		},
		{
			name: "remote oauth without oauth profile",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "oauth",
				cliProfile:   "work",
				oauthProfile: "",
				tokenFile:    tokenFile,
				cloudOrgID:   "",
				orgID:        "1234567",
			},
			want: "YANDEX_OAUTH_PROFILE",
		},
		{
			name: "remote iam without cli profile",
			env: acquisitionEnv{
				source:       "remote",
				tokenType:    "iam",
				cliProfile:   "",
				oauthProfile: "work",
				tokenFile:    "",
				cloudOrgID:   "cloud-org",
				orgID:        "",
			},
			want: "YANDEX_CLI_PROFILE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, value := range tt.env.variables() {
				t.Setenv(name, value)
			}

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
