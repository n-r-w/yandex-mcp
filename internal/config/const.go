package config

const (
	tokenSourceLocal  = "local"
	tokenSourceRemote = "remote"

	// TokenSourceOAuth selects a Yandex OAuth token read from a local file.
	TokenSourceOAuth = "oauth"

	headerCloudOrgID = "X-Cloud-Org-Id"
	headerOrgID      = "X-Org-Id"

	authSchemeBearer = "Bearer"
	authSchemeOAuth  = "OAuth"
)
