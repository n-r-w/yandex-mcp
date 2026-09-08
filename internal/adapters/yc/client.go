// Package yc obtains IAM tokens through the native Yandex Cloud CLI.
package yc

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/n-r-w/yandex-mcp/internal/adapters/ytoken"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// Source runs native yc. Callers coordinate shared acquisitions by profile.
type Source struct {
	path    string
	pattern *regexp.Regexp
}

var (
	_ ytoken.ITokenSource    = (*Source)(nil)
	_ authagent.ITokenSource = (*Source)(nil)
)

// New constructs a source for an executable path or local command name.
func New(path string) *Source {
	return &Source{path: path, pattern: regexp.MustCompile(tokenRegexPattern)}
}

// Acquire runs the native executable and waits for its exit before releasing the profile.
func (s *Source) Acquire(ctx context.Context, profile string) (string, error) {
	args := []string{"iam", "create-token"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	//nolint:gosec // executable is deployment configuration; arguments are fixed, no shell
	cmd := exec.CommandContext(ctx, s.path, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		if diagnostics.Len() > 0 {
			err = fmt.Errorf("%w\n%s", err, diagnostics.String())
		}
		return "", domain.AuthenticationError{Err: err}
	}
	if output.Len() == 0 {
		return "", domain.AuthenticationError{Err: errEmptyToken}
	}
	token := s.pattern.Find(output.Bytes())
	if token == nil {
		return "", domain.AuthenticationError{Err: errTokenNotFound}
	}
	return string(token), nil
}
