package authagent

import "time"

const (
	headerTimeout   = 5 * time.Second
	sshRetryDelay   = 5 * time.Second
	maxRequestBytes = 4096

	// Default OAuth profiles directory: <user config dir>/yandex-mcp/oauth.
	oauthDirName    = "yandex-mcp"
	oauthSubdirName = "oauth"
)
