// Package authremote requests workstation tokens through an SSH tunnel.
package authremote

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/adapters/ytoken"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// Source contacts only the configured loopback auth-agent endpoint.
type Source struct {
	tokenType string
	endpoint  string
	client    *http.Client
}

var (
	_ ytoken.ITokenSource    = (*Source)(nil)
	_ authagent.ITokenSource = (*Source)(nil)
)

// New constructs a remote source for one token type with proxies and redirects disabled.
func New(port int, tokenType string) (*Source, error) {
	address, err := config.AuthAgentAddress(port)
	if err != nil {
		return nil, err
	}
	//nolint:exhaustruct_v5 // no proxy and no independent request timeout
	client := &http.Client{
		Transport:     &http.Transport{},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	return &Source{tokenType: tokenType, endpoint: "http://" + address + "/token", client: client}, nil
}

// Acquire waits once for a token. Only the caller's context limits the wait.
func (s *Source) Acquire(ctx context.Context, profile string) (string, error) {
	token, err := s.request(ctx, profile)
	if err != nil {
		return "", domain.AuthenticationError{Err: err}
	}
	return token, nil
}

func (s *Source) request(ctx context.Context, profile string) (string, error) {
	if strings.TrimSpace(profile) == "" || s.tokenType == "" {
		return "", errInvalidRequest
	}
	// Released auth-agent versions reject unknown members, so an IAM request holds only the profile.
	requestType := s.tokenType
	if requestType == config.TokenTypeIAM {
		requestType = ""
	}
	body, err := json.Marshal(tokenRequest{TokenType: requestType, Profile: profile})
	if err != nil {
		return "", fmt.Errorf("%w: %w", errInvalidRequest, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %w", errInvalidRequest, err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("%w: %w", errUnavailable, err)
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := io.ReadAll(response.Body)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		return "", fmt.Errorf("%w: %w", errProtocol, err)
	}
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("%w: %w", errProtocol, err)
	}
	if contentType != "application/json" {
		return "", errProtocol
	}
	var result tokenResponse
	if err = json.Unmarshal(payload, &result, json.RejectUnknownMembers(true)); err != nil {
		return "", fmt.Errorf("%w: %w", errProtocol, err)
	}
	return result.result(response.StatusCode)
}

// Close releases idle HTTP connections.
func (s *Source) Close() { s.client.CloseIdleConnections() }
