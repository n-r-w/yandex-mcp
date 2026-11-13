package apihelpers

// HTTP headers and constants for Yandex API requests.
const (
	HeaderAuthorization = "Authorization"
	HeaderCloudOrgID    = "X-Cloud-Org-Id"
	HeaderContentType   = "Content-Type"

	ContentTypeJSON = "application/json"

	// AuthSchemeBearer is the default Authorization scheme, used for IAM tokens.
	AuthSchemeBearer = "Bearer"
)
