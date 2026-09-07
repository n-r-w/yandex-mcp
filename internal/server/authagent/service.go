// Package authagent serves workstation token acquisition over loopback HTTP.
package authagent

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"

	"github.com/n-r-w/yandex-mcp/internal/adapters/authwait"
)

// Service shares active acquisitions between remote MCP processes.
type Service struct {
	source       ITokenSource
	profiles     map[string]struct{}
	acquisitions authwait.Group
}

var _ http.Handler = (*Service)(nil)

// New constructs a handler for explicitly allowed profiles.
func New(source ITokenSource, profiles []string) *Service {
	allowed := make(map[string]struct{}, len(profiles))
	for _, profile := range profiles {
		allowed[profile] = struct{}{}
	}
	//nolint:exhaustruct_v5 // acquisitions start empty
	return &Service{source: source, profiles: allowed}
}

// ServeHTTP implements the synchronous token protocol on the trusted SSH tunnel.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/token" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		s.respond(w, http.StatusMethodNotAllowed, "", "invalid_request", "POST /token is required")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	var input tokenRequest
	if err != nil {
		s.respond(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if len(body) > maxRequestBytes {
		s.respond(w, http.StatusBadRequest, "", "invalid_request", "token request exceeds 4096 bytes")
		return
	}
	if err = json.Unmarshal(body, &input, json.RejectUnknownMembers(true)); err != nil {
		s.respond(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if input.Profile == "" {
		s.respond(w, http.StatusBadRequest, "", "invalid_request", "profile is required")
		return
	}
	if input.Version != protocolVersion {
		s.respond(w, http.StatusBadRequest, "", "incompatible_version", "incompatible auth-agent protocol version")
		return
	}
	if _, ok := s.profiles[input.Profile]; !ok {
		s.respond(w, http.StatusForbidden, "", "forbidden_profile", "profile access denied")
		return
	}
	token, err := s.acquisitions.Do(
		r.Context(),
		input.Profile,
		func(ctx context.Context) (string, error) { return s.source.Acquire(ctx, input.Profile) },
	)
	if r.Context().Err() != nil {
		return
	}
	if err != nil {
		s.respond(w, http.StatusBadGateway, "", "authentication_failed", err.Error())
		return
	}
	if token == "" {
		s.respond(w, http.StatusInternalServerError, "", "internal_error", "token source returned an empty token")
		return
	}
	s.respond(w, http.StatusOK, token, "", "")
}

func (s *Service) respond(w http.ResponseWriter, status int, token, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// A disconnected client needs no response retry. Never log the payload.
	_ = json.MarshalWrite(w, tokenResponse{Version: protocolVersion, Token: token, Error: code, Message: message})
}

// Close cancels owned acquisitions and waits for native process cleanup.
func (s *Service) Close() { s.acquisitions.Close() }
