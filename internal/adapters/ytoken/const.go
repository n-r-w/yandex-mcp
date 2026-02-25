package ytoken

// tokenRegexPattern matches Yandex IAM tokens in yc CLI output.
// Format: t1.[base64-like-chars][optional-padding].[86-base64-like-chars][optional-padding].
//
//nolint:gosec // G101: This is a regex pattern, not a hardcoded credential.
const tokenRegexPattern = `t1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{86}[=]{0,2}`

// ycCommandName is the only allowed CLI command for IAM token retrieval.
const ycCommandName = "yc"

// yc CLI fixed arguments for IAM token retrieval.
const (
	ycCommandArgIAM         = "iam"
	ycCommandArgCreateToken = "create-token"
	tokenRefreshRequestKey  = "token"
)
