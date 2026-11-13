package config

const (
	// TokenSourceLocal acquires tokens on the machine that runs MCP.
	TokenSourceLocal = "local"
	// TokenSourceRemote acquires tokens through the workstation auth-agent.
	TokenSourceRemote = "remote"

	// TokenTypeIAM selects IAM tokens, acquired through yc.
	TokenTypeIAM = "iam"
	// TokenTypeOAuth selects Yandex OAuth tokens, read from a file.
	TokenTypeOAuth = "oauth"

	// headerCloudOrgID carries the ID of a Yandex Cloud organization.
	headerCloudOrgID = "X-Cloud-Org-Id"
	// headerOrgID carries the ID of a Yandex 360 organization.
	headerOrgID = "X-Org-Id"

	// authSchemeBearer is the Authorization scheme of IAM tokens.
	authSchemeBearer = "Bearer"
	// authSchemeOAuth is the Authorization scheme of OAuth tokens.
	authSchemeOAuth = "OAuth"
)
