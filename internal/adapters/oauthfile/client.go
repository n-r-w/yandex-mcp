// Package oauthfile reads Yandex OAuth tokens from local files.
package oauthfile

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/n-r-w/yandex-mcp/internal/adapters/ytoken"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// Source reads one token file, or the file of the requested profile in a directory, at every acquisition.
type Source struct {
	path    string
	dir     string
	pattern *regexp.Regexp
}

var (
	_ ytoken.ITokenSource    = (*Source)(nil)
	_ authagent.ITokenSource = (*Source)(nil)
)

// New constructs a source that reads the absolute token file path for any requested profile.
func New(path string) *Source {
	return &Source{path: path, dir: "", pattern: nil}
}

// NewProfiles constructs a source that reads the file named after the requested profile in dir.
func NewProfiles(dir string) *Source {
	return &Source{path: "", dir: dir, pattern: regexp.MustCompile(profilePattern)}
}

// Acquire returns the token stored in the file selected by the source.
func (s *Source) Acquire(ctx context.Context, profile string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path, err := s.resolve(profile)
	if err != nil {
		return "", domain.AuthenticationError{Err: err}
	}
	//nolint:gosec // the path is deployment configuration; a profile name cannot leave its directory
	raw, err := os.ReadFile(path)
	if err != nil {
		if s.pattern != nil && errors.Is(err, fs.ErrNotExist) {
			err = fmt.Errorf("%w %q: %w", errUnknownProfile, profile, err)
		}
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

// resolve maps the profile to a file path. Profile names cannot leave the profiles directory.
func (s *Source) resolve(profile string) (string, error) {
	if s.pattern == nil {
		return s.path, nil
	}
	if s.dir == "" {
		return "", errNoProfilesDir
	}
	if !s.pattern.MatchString(profile) {
		return "", fmt.Errorf("%w: %q", errInvalidProfile, profile)
	}
	return filepath.Join(s.dir, profile), nil
}

// isForbiddenRune rejects characters that cannot appear in an HTTP header token value.
func isForbiddenRune(r rune) bool {
	return r > unicode.MaxASCII || unicode.IsSpace(r) || unicode.IsControl(r)
}
