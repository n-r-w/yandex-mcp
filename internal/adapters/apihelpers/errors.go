package apihelpers

import "fmt"

// HTTPError represents a non-2xx HTTP response.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

// Error implements the error interface for HTTPError.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, string(e.Body))
}
