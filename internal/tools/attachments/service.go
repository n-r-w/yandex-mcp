// Package attachments provides attachment delivery rules and disk writing for MCP tools.
package attachments

// Service holds the configured attachment delivery policy.
type Service struct {
	allowedExtensions []string
	allowedViewExts   []string
	allowedDirs       []string
}

// New creates a delivery service with the permitted disk extensions, inline extensions, and directories.
func New(allowedExtensions, allowedViewExts, allowedDirs []string) *Service {
	return &Service{
		allowedExtensions: normalizeAllowedExtensions(allowedExtensions),
		allowedViewExts:   normalizeAllowedExtensions(allowedViewExts),
		allowedDirs:       normalizeAllowedDirs(allowedDirs),
	}
}
