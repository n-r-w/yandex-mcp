package tracker

import "time"

const (
	headerAuthorization  = "Authorization"
	headerCloudOrgID     = "X-Cloud-Org-ID"
	headerContentType    = "Content-Type"
	headerAcceptLanguage = "Accept-Language"
	headerXTotalCount    = "X-Total-Count"
	headerXTotalPages    = "X-Total-Pages"
	headerXScrollID      = "X-Scroll-Id"
	headerXScrollToken   = "X-Scroll-Token" //nolint:gosec // not a credential
	headerLink           = "Link"

	contentTypeJSON = "application/json"
	acceptLangEN    = "en"

	defaultTimeout = 30 * time.Second
)
