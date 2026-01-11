package domain

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeBody_TruncatesLongContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "content shorter than limit",
			input:    "short content",
			maxLen:   100,
			expected: "short content",
		},
		{
			name:     "content exactly at limit",
			input:    "12345",
			maxLen:   5,
			expected: "12345",
		},
		{
			name:     "content exceeds limit",
			input:    "1234567890",
			maxLen:   5,
			expected: "12345",
		},
		{
			name:     "empty content",
			input:    "",
			maxLen:   100,
			expected: "",
		},
		{
			name:     "maxLen is zero returns empty string",
			input:    "some content",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "maxLen is negative returns empty string",
			input:    "some content",
			maxLen:   -1,
			expected: "",
		},
		{
			name:     "maxLen is very negative returns empty string",
			input:    "some content",
			maxLen:   -100,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBody(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeBody_LargeContent(t *testing.T) {
	t.Parallel()

	largeContent := strings.Repeat("x", 16*1024)
	result := SanitizeBody(largeContent, maxSanitizedBodySize)

	require.LessOrEqual(t, len(result), maxSanitizedBodySize)
	assert.Len(t, result, maxSanitizedBodySize)
}

func TestSanitizeBody_RemovesNonPrintableChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "null bytes removed",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "control chars removed",
			input:    "hello\x01\x02\x03world",
			expected: "helloworld",
		},
		{
			name:     "tabs preserved",
			input:    "hello\tworld",
			expected: "hello\tworld",
		},
		{
			name:     "newlines preserved",
			input:    "hello\nworld\r\n",
			expected: "hello\nworld\r\n",
		},
		{
			name:     "mixed content",
			input:    "line1\n\x00line2\t\x1fvalue",
			expected: "line1\nline2\tvalue",
		},
		{
			name:     "printable ASCII preserved",
			input:    "ABCabc123!@#$%^&*()",
			expected: "ABCabc123!@#$%^&*()",
		},
		{
			name:     "bell and backspace removed",
			input:    "hello\x07\x08world",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBody(tt.input, maxSanitizedBodySize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeBody_HandlesInvalidUTF8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid UTF-8 sequence",
			input: "hello\x80\x81world",
		},
		{
			name:  "truncated UTF-8 sequence",
			input: "hello\xc3",
		},
		{
			name:  "overlong UTF-8 encoding",
			input: "hello\xc0\x80world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBody(tt.input, maxSanitizedBodySize)
			assert.True(t, utf8.ValidString(result), "result must be valid UTF-8")
		})
	}
}

func TestSanitizeBody_PreservesValidUTF8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "cyrillic text",
			input:    "Привет мир",
			expected: "Привет мир",
		},
		{
			name:     "emoji",
			input:    "Hello 😀 World",
			expected: "Hello 😀 World",
		},
		{
			name:     "chinese characters",
			input:    "你好世界",
			expected: "你好世界",
		},
		{
			name:     "mixed unicode",
			input:    "Hello Привет 你好 😀",
			expected: "Hello Привет 你好 😀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBody(tt.input, maxSanitizedBodySize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpstreamError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      UpstreamError
		expected string
	}{
		{
			name: "with code",
			err: UpstreamError{
				Service:    ServiceWiki,
				Operation:  "get_page",
				HTTPStatus: 404,
				Code:       "PAGE_NOT_FOUND",
				Message:    "page not found",
				Details:    "",
			},
			expected: "wiki get_page: HTTP 404 (PAGE_NOT_FOUND): page not found",
		},
		{
			name: "without code",
			err: UpstreamError{
				Service:    ServiceTracker,
				Operation:  "get_issue",
				HTTPStatus: 500,
				Code:       "",
				Message:    "internal error",
				Details:    "",
			},
			expected: "tracker get_issue: HTTP 500: internal error",
		},
		{
			name: "with details",
			err: UpstreamError{
				Service:    ServiceWiki,
				Operation:  "list_pages",
				HTTPStatus: 400,
				Code:       "",
				Message:    "bad request",
				Details:    "invalid cursor format",
			},
			expected: "wiki list_pages: HTTP 400: bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewUpstreamError(t *testing.T) {
	t.Parallel()

	rawBody := strings.Repeat("x", 10*1024)
	err := NewUpstreamError(ServiceTracker, "search_issues", 500, "", "server error", rawBody)

	assert.Equal(t, ServiceTracker, err.Service)
	assert.Equal(t, "search_issues", err.Operation)
	assert.Equal(t, 500, err.HTTPStatus)
	assert.Empty(t, err.Code)
	assert.Equal(t, "server error", err.Message)
	require.LessOrEqual(t, len(err.Details), maxSanitizedBodySize)
}
