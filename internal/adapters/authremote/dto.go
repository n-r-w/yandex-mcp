package authremote

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type tokenRequest struct {
	Version int    `json:"version"`
	Profile string `json:"profile"`
}

type tokenResponse struct {
	Version int    `json:"version"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r tokenResponse) result(status int) (string, error) {
	if r.Version != protocolVersion {
		return "", errIncompatible
	}
	if status == http.StatusOK && r.Token != "" && r.Error == "" && r.Message == "" &&
		!strings.ContainsAny(r.Token, " \t\r\n") {
		return r.Token, nil
	}
	if r.Token != "" || r.Message == "" {
		return "", errProtocol
	}
	cause := r.errorCause(status)
	if errors.Is(cause, errProtocol) {
		return "", errProtocol
	}
	return "", fmt.Errorf("%w: %s", cause, r.Message)
}

func (r tokenResponse) errorCause(status int) error {
	var cause error
	switch {
	case status == http.StatusBadRequest && r.Error == "incompatible_version":
		cause = errIncompatible
	case status == http.StatusBadRequest && r.Error == "invalid_request":
		cause = errInvalidRequest
	case status == http.StatusForbidden && r.Error == "forbidden_profile":
		cause = errForbidden
	case status == http.StatusBadGateway && r.Error == "authentication_failed":
		cause = errAuthentication
	case status == http.StatusInternalServerError && r.Error == "internal_error":
		cause = errInternal
	default:
		return errProtocol
	}
	return cause
}
