package wiki

import "time"

const (
	headerAuthorization = "Authorization"
	headerCloudOrgID    = "X-Cloud-Org-Id"
	headerContentType   = "Content-Type"

	contentTypeJSON = "application/json"

	defaultTimeout   = 30 * time.Second
	maxResourcesSize = 50
	maxGridsSize     = 50
)
