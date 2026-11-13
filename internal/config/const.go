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

	headerCloudOrgID = "X-Cloud-Org-Id"
	headerOrgID      = "X-Org-Id"

	authSchemeBearer = "Bearer"
	authSchemeOAuth  = "OAuth"
)
