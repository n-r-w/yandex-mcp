package apihelpers

import "time"

// HTTP headers and constants for Yandex API requests.
const (
	HeaderAuthorization = "Authorization"
	HeaderCloudOrgID    = "X-Cloud-Org-Id"
	HeaderContentType   = "Content-Type"

	ContentTypeJSON = "application/json"

	DefaultTimeout = 30 * time.Second
)
