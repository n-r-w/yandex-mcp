// Package domain defines core types and errors shared across the application.
package domain

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxSanitizedBodySize is the maximum size of sanitized upstream response body.
// Limited to prevent large error messages in logs and responses.
const MaxSanitizedBodySize = 4 * 1024 // 4 KiB

// httpStatusUnauthorized is the HTTP status code for unauthorized requests.
const httpStatusUnauthorized = 401

// Service represents the upstream Yandex service.
type Service string

// Service constants for upstream Yandex services.
const (
	ServiceWiki    Service = "wiki"
	ServiceTracker Service = "tracker"
)

// UpstreamError represents an error from Yandex upstream APIs.
// It contains sanitized details suitable for logging and error responses.
type UpstreamError struct {
	Service    Service
	Operation  string
	HTTPStatus int
	Code       string // optional, e.g. error_code from Wiki API
	Message    string // safe, short description
	Details    string // optional, sanitized body snippet
}

// Error implements the error interface.
func (e UpstreamError) Error() string {
	var b strings.Builder
	b.WriteString(string(e.Service))
	b.WriteString(" ")
	b.WriteString(e.Operation)
	b.WriteString(": HTTP ")
	b.WriteString(strconv.Itoa(e.HTTPStatus))

	if e.Code != "" {
		b.WriteString(" (")
		b.WriteString(e.Code)
		b.WriteString(")")
	}

	b.WriteString(": ")
	b.WriteString(e.Message)

	return b.String()
}

// IsRetryable returns true if the error indicates a condition where
// retrying the request (with token refresh) may succeed.
// Currently only HTTP 401 triggers retry with token refresh.
func (e UpstreamError) IsRetryable() bool {
	return e.HTTPStatus == httpStatusUnauthorized
}

// NewUpstreamError creates a new UpstreamError with sanitized details.
func NewUpstreamError(
	service Service,
	operation string,
	httpStatus int,
	code string,
	message string,
	rawBody string,
) UpstreamError {
	return UpstreamError{
		Service:    service,
		Operation:  operation,
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Details:    SanitizeBody(rawBody, MaxSanitizedBodySize),
	}
}

// SanitizeBody sanitizes upstream response body for safe logging and error messages.
// It:
// - Truncates content to maxLen bytes
// - Removes non-printable characters (except tab, newline, carriage return)
// - Ensures the result is valid UTF-8.
func SanitizeBody(body string, maxLen int) string {
	if body == "" || maxLen <= 0 {
		return ""
	}

	// First, make the string valid UTF-8 by replacing invalid sequences
	body = ensureValidUTF8(body)

	// Remove non-printable characters
	body = removeForbiddenChars(body)

	// Truncate to maxLen
	if len(body) > maxLen {
		body = truncateToValidUTF8(body, maxLen)
	}

	return body
}

// ensureValidUTF8 replaces invalid UTF-8 sequences with empty string.
func ensureValidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Skip invalid byte
			i++
			continue
		}
		b.WriteRune(r)
		i += size
	}

	return b.String()
}

// removeForbiddenChars removes control characters except tab, newline, and carriage return.
func removeForbiddenChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		if isPrintableOrAllowed(r) {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// isPrintableOrAllowed returns true if the rune should be kept in sanitized output.
func isPrintableOrAllowed(r rune) bool {
	// Allow tab, newline, carriage return
	if r == '\t' || r == '\n' || r == '\r' {
		return true
	}

	// Remove all other control characters
	if unicode.IsControl(r) {
		return false
	}

	// Keep printable characters (including non-ASCII unicode)
	return unicode.IsPrint(r) || unicode.IsSpace(r)
}

// truncateToValidUTF8 truncates s to at most maxLen bytes,
// ensuring the result is valid UTF-8 (doesn't cut in the middle of a rune).
func truncateToValidUTF8(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Find the last valid UTF-8 boundary at or before maxLen
	for maxLen > 0 && !utf8.RuneStart(s[maxLen]) {
		maxLen--
	}

	return s[:maxLen]
}
