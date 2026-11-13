// Package oauthfile reads a Yandex OAuth token from a local file.
package oauthfile

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/n-r-w/yandex-mcp/internal/adapters/ytoken"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// Source reads the token file on every acquisition, so a replaced token is used without a restart.
type Source struct {
	path string
}

var (
	_ ytoken.ITokenSource    = (*Source)(nil)
	_ authagent.ITokenSource = (*Source)(nil)
)

// New constructs a source for an absolute token file path.
func New(path string) *Source {
	return &Source{path: path}
}

// Acquire returns the token stored in the file. The profile is unused: the file identifies the account.
func (s *Source) Acquire(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return "", domain.AuthenticationError{Err: fmt.Errorf("read OAuth token file: %w", err)}
	}
	// Text editors may prepend a byte order mark and append a line break.
	token := strings.TrimSpace(strings.TrimPrefix(string(raw), byteOrderMark))
	if token == "" {
		return "", domain.AuthenticationError{Err: errEmptyToken}
	}
	if strings.ContainsFunc(token, isForbiddenRune) {
		return "", domain.AuthenticationError{Err: errMalformedToken}
	}
	return token, nil
}

// isForbiddenRune rejects characters that cannot appear in an HTTP header token value.
func isForbiddenRune(r rune) bool {
	return r > unicode.MaxASCII || unicode.IsSpace(r) || unicode.IsControl(r)
}
