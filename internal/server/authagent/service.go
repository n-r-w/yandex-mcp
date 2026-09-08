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
	acquisitions authwait.Group
}

var _ http.Handler = (*Service)(nil)

// New constructs the workstation token handler.
func New(source ITokenSource) *Service {
	//nolint:exhaustruct_v5 // acquisitions start empty
	return &Service{source: source}
}

// ServeHTTP waits for the requested profile's token or request cancellation.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/token" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		s.respond(w, http.StatusMethodNotAllowed, "", "POST /token is required")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil {
		s.respond(w, http.StatusBadRequest, "", err.Error())
		return
	}
	if len(body) > maxRequestBytes {
		s.respond(w, http.StatusBadRequest, "", "token request exceeds 4096 bytes")
		return
	}
	var input tokenRequest
	if err = json.Unmarshal(body, &input, json.RejectUnknownMembers(true)); err != nil {
		s.respond(w, http.StatusBadRequest, "", err.Error())
		return
	}
	if input.Profile == "" {
		s.respond(w, http.StatusBadRequest, "", "profile is required")
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
		s.respond(w, http.StatusBadGateway, "", err.Error())
		return
	}
	if token == "" {
		s.respond(w, http.StatusInternalServerError, "", "token source returned an empty token")
		return
	}
	s.respond(w, http.StatusOK, token, "")
}

// respond preserves diagnostics without logging successful token output.
func (s *Service) respond(w http.ResponseWriter, status int, token, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// A disconnected client needs no response retry. Never log the payload.
	_ = json.MarshalWrite(w, tokenResponse{Token: token, Message: message})
}

// Close cancels acquisitions and waits for native process cleanup.
func (s *Service) Close() { s.acquisitions.Close() }
